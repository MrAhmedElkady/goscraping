package fetch

import (
	"bytes"
	"context"
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

	utls "github.com/refraction-networking/utls"
)

// Session represents a persistent browser session
type Session struct {
	ID     string
	Client *http.Client
	Jar    *cookies.SessionJar
	// We could store the assigned fingerprint here to ensure stability
	FingerprintID utls.ClientHelloID
}

var (
	sessionStore = make(map[string]*Session)
	storeMu      sync.RWMutex
)

// getOrCreateSession retrieves or creates a session
func getOrCreateSession(id string, timeout time.Duration) (*Session, error) {
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

	// Assign a stable fingerprint for this session
	// For now, randomly pick between Chrome and Safari? or Default to Chrome.
	// Real implementation: randomize once, persist.
	fpID := client.FingerprintChrome120
	if time.Now().UnixNano()%2 == 0 {
		// fpID = client.FingerprintSafari16 // Optional randomization
	}

	transport := client.NewTransport(fpID)

	httpClient := &http.Client{
		Transport: transport,
		Jar:       jar,
		Timeout:   timeout,
	}

	sess := &Session{
		ID:            id,
		Client:        httpClient,
		Jar:           jar,
		FingerprintID: fpID,
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

	// 1. Get Session (Transport is reused)
	sess, err := getOrCreateSession(opts.SessionID, opts.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to init session: %w", err)
	}

	// 2. Prepare Request
	req, err := http.NewRequest(opts.Method, urlStr, bytes.NewBuffer(opts.Body))
	if err != nil {
		return nil, err
	}

	// 3. Generate & Apply Headers
	// Map Options.HeaderConfig to headers.Config
	hConfig := headers.Config{
		Browser: opts.HeaderConfig.Browser,
		OS:      opts.HeaderConfig.OS,
		Device:  opts.HeaderConfig.Device,
		Locale:  opts.HeaderConfig.Locale,
	}

	// If user didn't specify, maybe pull from session?
	// For now, use options or defaults.
	browserHeaders := headers.Generate(hConfig)
	for k, v := range browserHeaders {
		// Only set if not already set? Or overwrite?
		// Usually browser headers are base.
		req.Header[k] = v
	}

	// Apply user overrides
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	// 4. Proxy Selection
	currentProxy := opts.ProxyURL
	proxyList := opts.Proxies

	// Helper to rotate proxy
	rotateProxy := func() {
		if len(proxyList) > 0 {
			// Simple rotation: pop from front, append to back?
			// Or just random pick?
			// Let's just pick next.
			currentProxy = proxyList[0]
			if len(proxyList) > 1 {
				proxyList = append(proxyList[1:], proxyList[0])
			}
		}
	}

	// 5. Execute with Retry
	var httpResp *http.Response
	policy := retry.DefaultPolicy()

	for i := 0; i <= policy.MaxAttempts; i++ {
		// Set Proxy in Context
		ctx := req.Context()
		if currentProxy != "" {
			ctx = context.WithValue(ctx, client.CtxProxyURL, currentProxy)
		}
		reqWithCtx := req.WithContext(ctx)

		httpResp, err = sess.Client.Do(reqWithCtx)
		if err != nil {
			// Network error
			if i < policy.MaxAttempts {
				rotateProxy()
				continue
			}
			return nil, err
		}

		if policy.ShouldRetry(httpResp.StatusCode) {
			httpResp.Body.Close()
			if i < policy.MaxAttempts {
				rotateProxy()
				time.Sleep(1 * time.Second) // Backoff
				continue
			}
		}

		// Success
		break
	}

	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	// 6. Read Body
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
