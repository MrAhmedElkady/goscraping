package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	utls "github.com/refraction-networking/utls"
)

// ContextKey type for context values
type ContextKey string

const (
	// CtxProxyURL is the context key for the proxy URL
	CtxProxyURL ContextKey = "proxy_url"
)

// TransportManager implements http.RoundTripper with uTLS support.
// It manages a persistent http.Transport to ensure connection reuse.
// It explicitly enforces HTTP/1.1 to match uTLS capabilities safely.
type TransportManager struct {
	// ClientHelloID is the uTLS ClientHello ID to use.
	ClientHelloID utls.ClientHelloID

	// underlying is the actual http.Transport that manages connections.
	// We create this ONCE and reuse it.
	underlying *http.Transport

	// ForceHTTP1 dictates whether we strictly enforce HTTP/1.1 in ALPN.
	// Default: true (safest for generic scraping).
	ForceHTTP1 bool
}

// NewTransportManager creates a new persistent transport manager.
// This is the V2 recommended constructor.
func NewTransportManager(clientHello utls.ClientHelloID) *TransportManager {
	t := &TransportManager{
		ClientHelloID: clientHello,
		ForceHTTP1:    true,
	}

	// Create the persistent http.Transport ONCE.
	// We hook DialTLSContext to inject uTLS logic.
	t.underlying = &http.Transport{
		DialTLSContext: t.dialTLS,
		// DialContext is used for plain HTTP.
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2: false, // We control H2 via ALPN manually
		MaxIdleConns:      100,
		IdleConnTimeout:   90 * time.Second,
		// We do NOT set Proxy here because we handle specific proxy logic in DialTLSContext
		// to ensure uTLS works correctly over the proxy tunnel.
		Proxy: nil,
	}

	return t
}

// RoundTrip delegates to the persistent underlying Transport.
func (t *TransportManager) RoundTrip(req *http.Request) (*http.Response, error) {
	// Check for proxy in context for plain HTTP
	// For HTTPS, dialTLS handles it.
	// For HTTP, we might need to handle it if we want proxy support there too.
	// Current implementation keeps it simple for HTTPS focus.
	return t.underlying.RoundTrip(req)
}

// dialTLS handles the TCP connection (possibly via Proxy) and the uTLS handshake.
func (t *TransportManager) dialTLS(ctx context.Context, network, addr string) (net.Conn, error) {
	// 1. Determine destination host/port
	hostname := getServerName(addr)

	// 2. Check for Proxy in Context
	var rawConn net.Conn
	var err error

	proxyURL, ok := ctx.Value(CtxProxyURL).(string)

	if ok && proxyURL != "" {
		// Proxy Path: Dial Proxy -> CONNECT -> uTLS
		pd, err := MakeProxyDialer(proxyURL, 30*time.Second)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy: %w", err)
		}
		// This 'dialer' already does the CONNECT handshake if it's an HTTP proxy
		// or SOCKS5 handshake as appropriate.
		rawConn, err = pd(ctx, network, addr)
		if err != nil {
			return nil, err // Proxy connection failed
		}
	} else {
		// Direct Path: TCP Dial -> uTLS
		dialer := &net.Dialer{Timeout: 30 * time.Second}
		rawConn, err = dialer.DialContext(ctx, network, addr)
		if err != nil {
			return nil, err
		}
	}

	// 3. Prepare uTLS connection with ALPN patching
	// uTLS Presets (like HelloChrome_120) HARDCODE the ALPN extension to include "h2".
	// utls.Config.NextProtos DOES NOT override this for Presets.
	// We must manually edit the Spec to enforce HTTP/1.1 if desired.

	var uConn *utls.UConn

	if t.ForceHTTP1 {
		// 3a. Get the base spec for the Fingerprint
		spec, err := utls.UTLSIdToSpec(t.ClientHelloID)
		if err != nil {
			rawConn.Close()
			return nil, fmt.Errorf("failed to get uTLS spec: %w", err)
		}

		// 3b. Patch ALPN extension to remove "h2"
		for i, ext := range spec.Extensions {
			if _, ok := ext.(*utls.ALPNExtension); ok {
				// We found the ALPN extension. Replace it with just http/1.1
				// Note: We modify the spec slice directly.
				newAlpn := &utls.ALPNExtension{AlpnProtocols: []string{"http/1.1"}}
				spec.Extensions[i] = newAlpn
				break
			}
		}

		// 3c. Create Conn with Custom Spec
		// We use HelloCustom because we are providing a modified spec
		uConn = utls.UClient(rawConn, &utls.Config{
			ServerName: hostname,
			// NextProtos here is a fallback for HelloCustom but good to keep consistent
			NextProtos: []string{"http/1.1"},
			MinVersion: tls.VersionTLS12,
		}, utls.HelloCustom)

		// Apply the spec
		if err := uConn.ApplyPreset(&spec); err != nil {
			rawConn.Close()
			return nil, fmt.Errorf("failed to apply uTLS spec: %w", err)
		}

	} else {
		// Standard behavior (might negotiate h2)
		uConn = utls.UClient(rawConn, &utls.Config{
			ServerName: hostname,
			NextProtos: []string{"h2", "http/1.1"},
			MinVersion: tls.VersionTLS12,
		}, t.ClientHelloID)
	}

	// 4. Handshake
	if err := uConn.Handshake(); err != nil {
		rawConn.Close()
		return nil, fmt.Errorf("uTLS handshake failed: %w", err)
	}

	// 5. Verify Negotiated Protocol (Optional Debugging)

	return uConn, nil
}

func getServerName(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}
