package client

import (
	"context"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/http2"
)

// Transport implements http.RoundTripper with uTLS support.
type Transport struct {
	// ProxyURL is the URL of the proxy to use.
	ProxyURL string

	// ClientHelloID is the uTLS ClientHello ID to use.
	ClientHelloID utls.ClientHelloID

	// cachedTransports stores per-host transports if needed (mostly for h2 persistence,
	// although for scraping rotation is common. For now we use a single transport logic per request style or shared?)
	// To match actual browser behavior, we usually need a dialer that handles the handshake.

	// We will use a custom Dialer.
	Dialer *net.Dialer
}

// NewTransport creates a new Transport with the given options.
func NewTransport(clientHello utls.ClientHelloID) *Transport {
	return &Transport{
		ClientHelloID: clientHello,
		Dialer:        &net.Dialer{Timeout: 30 * time.Second}, // Default timeout
	}
}

// RoundTrip executes a single HTTP transaction.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// For now, we are creating a fresh connection or utilizing a pool conceptually.
	// If we want FULL control over the ClientHello, we often have to dial manually for HTTPS.

	// However, Go's http2.Transport doesn't interact easily with uTLS net.Conn directly
	// without some work.

	// We will implement a basic version first that handles HTTP/1.1 and initiates HTTP/2 if negotiated.

	// This is a simplified implementation. Real-world robust uTLS usage with HTTP/2 often requires
	// hooking into the DialTLSContext of http.Transport.

	// Let's delegate to a proper http.Transport with a custom DialTLSContext.

	rt := &http.Transport{
		DialTLSContext:    t.dialTLS,
		DialContext:       t.Dialer.DialContext,
		ForceAttemptHTTP2: true,
	}

	// We likely want to configure HTTP/2 transport explicitly to match settings.
	h2t, err := http2.ConfigureTransports(rt)
	if err == nil {
		// Here we can tune h2t settings if exposed, but standard library limits what we can tune
		// on the global transport level easily without forking.
		// For "Tune SETTINGS", we might need to send raw frames or use a specialized library
		// if generic http2 is insufficient. But usually standard http2 is "okay" if TLS is right,
		// though strict parity needs more.

		// Note: request wants "Tune SETTINGS, WINDOW_SIZE".
		// golang.org/x/net/http2 exposes some of this.
		h2t.InitialWindowSize = 6291456 // Example chrome-ish
		h2t.HeaderTableSize = 65536
		h2t.PushHandler = nil
	}

	return rt.RoundTrip(req)
}

func (t *Transport) dialTLS(ctx context.Context, network, addr string) (net.Conn, error) {
	// 1. Dial TCP
	rawConn, err := t.Dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	// 2. Wrap with uTLS
	uConn := utls.UClient(rawConn, &utls.Config{
		ServerName: getServerName(addr),
		// InsecureSkipVerify: true, // Optional: might want to expose this
		NextProtos: []string{"h2", "http/1.1"},
	}, t.ClientHelloID)

	// 3. Handshake
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
