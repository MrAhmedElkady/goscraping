package client

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

// ContextKey type for context values
type ContextKey string

const (
	// CtxProxyURL is the context key for the proxy URL
	CtxProxyURL ContextKey = "proxy_url"
)

// Transport implements http.RoundTripper with uTLS support.
type Transport struct {
	// ClientHelloID is the uTLS ClientHello ID to use.
	ClientHelloID utls.ClientHelloID

	// DialContext specifies the dial function for creating TCP connections.
	// If nil, &net.Dialer{}.DialContext is used.
	// This dialer can be wrapped to support proxies.
	DialContext func(ctx context.Context, network, addr string) (net.Conn, error)
}

// NewTransport creates a new Transport.
func NewTransport(clientHello utls.ClientHelloID) *Transport {
	// Default dialer
	dialer := &net.Dialer{Timeout: 30 * time.Second}
	return &Transport{
		ClientHelloID: clientHello,
		DialContext:   dialer.DialContext,
	}
}

// RoundTrip executes a single HTTP transaction.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// We utilize the underlying http.Transport for connection management and HTTP/2 support.
	// Crucially, we hook into DialTLSContext to perform the uTLS handshake.

	// Check for proxy in context to determine if we need to wrap the dialer for this request.
	// However, http.Transport caches connections based on (Scheme, Host, Proxy).
	// If we use varying proxies per request, we need to ensure http.Transport uses the correct proxy.
	// http.Transport has a Proxy field: func(*Request) (*url.URL, error).

	// Strategy:
	// 1. Define a Proxy function that reads from the Request Context.
	// 2. Pass this function to http.Transport.
	// 3. For HTTPS (DialTLSContext), we essentially perform "CONNECT" if a proxy is present, then uTLS.
	//    BUT http.Transport with a Proxy function set AUTOMATICALLY handles the CONNECT
	//    before calling DialTLSContext??
	//    Let's check docs/behavior:
	//    - If Proxy is set, http.Transport connects to the proxy.
	//    - Then it calls DialTLSContext. The 'network'/'addr' passed to DialTLSContext
	//      depends. Usually, for HTTPS through proxy, it dials the proxy, does CONNECT...
	//      ACTUALLY: DialTLSContext is specifically for the *TLS handshake* on an *already connected* socket?
	//      No, DialTLSContext is "Dial + Handshake".
	//    - If generic Proxy is used, http.Transport does the TCP dial to proxy + CONNECT.
	//      Then it expects a TLS handshake.
	//      However, it doesn't expose the "underlying connection to proxy" to DialTLSContext easily
	//      if it handles the CONNECT itself. It usually expects DialTLSContext to do EVERYTHING (Dial+Handshake).

	//    - Therefore, if we define DialTLSContext, WE are responsible for traversing the proxy!
	//      http.Transport will NOT do the CONNECT for us if DialTLSContext is set (mostly).

	//    - So, we will implement the proxy traversal inside dialTLS.

	rt := &http.Transport{
		DialTLSContext:    t.dialTLS,
		DialContext:       t.DialContext, // Used for HTTP (plain)
		ForceAttemptHTTP2: true,
		// We do NOT set Proxy here because we handle it manually in DialTLS/Dial.
		// If we set Proxy, standard transport might try to do things we don't want,
		// or it's redundant if our Dial functions handle it.
		// Exception: For plain HTTP requests, we might want standard Proxy support?
		// Let's stick to doing it in the Dialer for consistency.
		Proxy: nil,
	}

	if _, err := http2.ConfigureTransports(rt); err != nil {
		// ignore error if h2 not supported or already configured
	}

	return rt.RoundTrip(req)
}

func (t *Transport) dialTLS(ctx context.Context, network, addr string) (net.Conn, error) {
	// 1. Determine dialer (Direct vs Proxy)
	dialer := t.DialContext

	proxyURL, ok := ctx.Value(CtxProxyURL).(string)
	if ok && proxyURL != "" {
		// Create a proxy dialer for this request
		// We use the helper from proxy.go
		// Note: MakeProxyDialer returns a dial function that handles SOCKS5 or HTTP CONNECT
		pd, err := MakeProxyDialer(proxyURL, 30*time.Second)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy: %w", err)
		}
		dialer = pd
	}

	// 2. Dial TCP (through proxy if configured)
	rawConn, err := dialer(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	// 3. Wrap with uTLS
	uConn := utls.UClient(rawConn, &utls.Config{
		ServerName: getServerName(addr),
		NextProtos: []string{"h2", "http/1.1"},
		// InsecureSkipVerify: true, // TODO: Make configurable?
	}, t.ClientHelloID)

	// 4. Handshake
	if err := uConn.Handshake(); err != nil {
		_ = rawConn.Close()
		return nil, err
	}

	return uConn, nil
}
func getServerName(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}
