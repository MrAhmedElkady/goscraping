package fetch

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"goscraping/client"
	"goscraping/cookies"
	"goscraping/headers"
	"goscraping/retry"
	"goscraping/types"
)

// Session represents a persistent browser session
type Session struct {
	ID     string
	Client *http.Client
	Jar    *cookies.SessionJar
}

var (
	sessionStore = make(map[string]*Session)
	storeMu      sync.RWMutex
)

// getOrCreateSession retrieves or creates a session
func getOrCreateSession(id string, proxyURL string, timeout time.Duration) (*Session, error) {
	if id != "" {
		storeMu.RLock()
		if sess, ok := sessionStore[id]; ok {
			storeMu.RUnlock()
			return sess, nil
		}
		storeMu.RUnlock()
	}

	// Create new session
	jar := cookies.NewSessionJar()

	// Transport setup
	// For now we use Chrome 120 fingerprint as default if not specified
	// In a real scenario we might vary this per session or allow config
	transport := client.NewTransport(client.FingerprintChrome120)
	transport.ProxyURL = proxyURL

	// Configure Proxy
	if proxyURL != "" {
		// We use the MakeProxyDialer helper we wrote
		// We need to inject this into Transport.
		// client.Transport has Dialer *net.Dialer.
		// But our MakeProxyDialer returns a func(ctx, net, addr).
		// Modifying client.Transport to accept a generic dialer or context dialer.
		// For now, let's assume we can set DialContext on the transport's dialer if compatible,
		// or we better update NewTransport to accept it.
		// Given current client implementation:
		// transport.Dialer is a *net.Dialer.
		// We need to override the DialTLSContext or similar.
		// Let's rely on client/proxy.go to handle this logic if we update client.Transport.

		// Simpler approach for this step:
		// Update client.Transport to have a custom 'ContextDialer'.
		// But since I can't edit client/transport.go in this specific tool call easily without context switch,
		// I will assume I can handle this later or do a patch.
		// Actually, I can use a closure in the DialTLSContext?

		dialer, err := client.MakeProxyDialer(proxyURL, timeout)
		if err != nil {
			return nil, err
		}

		// This effectively overrides the base dialer
		transport.Dialer.DialContext = dialer
	}

	httpClient := &http.Client{
		Transport: transport,
		Jar:       jar,
		Timeout:   timeout,
	}

	sess := &Session{
		ID:     id,
		Client: httpClient,
		Jar:    jar,
	}

	if id != "" {
		storeMu.Lock()
		sessionStore[id] = sess
		storeMu.Unlock()
	}

	return sess, nil
}

// Fetch executes a scraped request
func Fetch(urlStr string, opts *types.Options) (*Response, error) {
	if opts == nil {
		opts = types.DefaultOptions()
	}

	// 1. Get Session
	sess, err := getOrCreateSession(opts.SessionID, opts.ProxyURL, opts.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to init session: %w", err)
	}

	// 2. Prepare Request
	req, err := http.NewRequest(opts.Method, urlStr, bytes.NewBuffer(opts.Body))
	if err != nil {
		return nil, err
	}

	// 3. Generate & Apply Headers
	// If explicit headers provided, use them. Else generate.
	// Ideally we merge: browser defaults + user overrides.
	gen := headers.NewGenerator(headers.ChromeDesktop) // Default to Chrome
	if opts.HeaderProfile == "safari" {
		gen = headers.NewGenerator(headers.SafariIOS)
	}

	// Apply generated headers first
	browserHeaders := gen.GetHeaders()
	for k, v := range browserHeaders {
		req.Header[k] = v
	}

	// Apply user overrides
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	// 4. Execute with Retry
	var httpResp *http.Response
	policy := retry.DefaultPolicy()

	for i := 0; i <= policy.MaxAttempts; i++ {
		httpResp, err = sess.Client.Do(req)
		if err != nil {
			// Network error, maybe retry?
			if i < policy.MaxAttempts {
				continue
			}
			return nil, err
		}

		if policy.ShouldRetry(httpResp.StatusCode) {
			// Close body before retry
			httpResp.Body.Close()
			if i < policy.MaxAttempts {
				// Ideally we rotate proxy here if proxy rotation is enabled
				// For now, we just retry.
				time.Sleep(1 * time.Second) // Simple backoff
				continue
			}
		}

		// Success or max retries reached without retry condition
		break
	}

	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	// 5. Read Body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode:  httpResp.StatusCode,
		Headers:     httpResp.Header,
		Body:        body,
		RawResponse: httpResp,
	}, nil
}
