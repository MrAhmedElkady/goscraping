package client

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

// MakeProxyDialer returns a DialContext function that routes through the given proxy URL.
// Supports socks5, http, and https schemes.
func MakeProxyDialer(proxyURL string, timeout time.Duration) (func(context.Context, string, string) (net.Conn, error), error) {
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	var dialer proxy.Dialer
	baseDialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second,
	}

	switch parsedURL.Scheme {
	case "socks5", "socks5h":
		dialer, err = proxy.FromURL(parsedURL, baseDialer)
		if err != nil {
			return nil, err
		}
	case "http", "https":
		// For HTTP/HTTPS proxies, net/http Transport handles it via Proxy field usually.
		// Use proxy.FromURL doesn't support http/https for raw dialing typically in the same way
		// without a CONNECT implementation.
		// However, since we are using utls which wraps a connection, we need a transparent TCP connection
		// through the proxy (CONNECT method).
		// A simple "http proxy dialer" is needed if we are doing raw TLS handshake over it.
		// There isn't a standard 'dialer' for HTTP proxies in x/net/proxy that does CONNECT.
		// We might need a small helper or rely on the Transport's built-in Proxy support
		// BUT utls needs a raw conn.

		// This is a known complexity with utls + http proxy. `http.Transport` does the CONNECT
		// automatically if you provide `Proxy` field, BUT it then does the TLS handshake itself.
		// To use uTLS, we must do the CONNECT ourselves.
		return newHTTPConnectDialer(parsedURL, baseDialer), nil

	default:
		return baseDialer.DialContext, nil
	}

	// Wrap proxy.Dialer to ContextDialer
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		// x/net/proxy Dialer interface is Dial(network, addr)
		// Some implementations support DialContext, but the interface definition is Dial.
		// We can cast if supported, or just use Dial with a goroutine for context (simplified here).
		if d, ok := dialer.(proxy.ContextDialer); ok {
			return d.DialContext(ctx, network, addr)
		}
		return dialer.Dial(network, addr) // Context ignored if not supported
	}, nil
}

// newHTTPConnectDialer returns a dialer that performs HTTP CONNECT
func newHTTPConnectDialer(proxyURL *url.URL, forward *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		// 1. Dial the proxy
		conn, err := forward.DialContext(ctx, "tcp", proxyURL.Host)
		if err != nil {
			return nil, err
		}

		// 2. Send CONNECT request
		// Simplified basic auth handling
		req := &http.Request{
			Method: "CONNECT",
			URL:    &url.URL{Opaque: addr},
			Host:   addr,
			Header: make(http.Header),
		}
		if proxyURL.User != nil {
			password, _ := proxyURL.User.Password()
			auth := proxyURL.User.Username() + ":" + password
			basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
			req.Header.Set("Proxy-Authorization", basicAuth)
		}

		if err := req.Write(conn); err != nil {
			conn.Close()
			return nil, err
		}

		// 3. Read response
		resp, err := http.ReadResponse(bufio.NewReader(conn), req)
		if err != nil {
			conn.Close()
			return nil, err
		}
		if resp.StatusCode != 200 {
			conn.Close()
			return nil, fmt.Errorf("proxy connection failed: %s", resp.Status)
		}

		return conn, nil
	}
}
