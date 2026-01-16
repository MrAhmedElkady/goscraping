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
	"goscraping/identity"
	"goscraping/retry"
	"goscraping/types"
)

// Session represents a persistent browser session
type Session struct {
	ID     string
	Client *http.Client
	Jar    *cookies.SessionJar

	// The specific Identity assigned to this session
	Identity identity.Identity

	// Mutex for safety
	mu sync.RWMutex
}

var (
	sessionStore = make(map[string]*Session)
	storeMu      sync.RWMutex
)

// getOrCreateSession retrieves or creates a session
func getOrCreateSession(id string, config types.IdentityConfig, timeout time.Duration) (*Session, error) {
	// 1. Try to retrieve existing session
	if id != "" {
		storeMu.RLock()
		sess, ok := sessionStore[id]
		storeMu.RUnlock()
		if ok {
			// Check if we need to force rotate identity (Randomize=true AND unstable?)
			// If Stable=true, we keep it. If Randomize=true, we might want to rotate.
			// Requirement: "Reuse that Identity consistently ... unless rotated explicitly"
			// So default behavior is reuse.
			// If user passed Randomize=true in this request options, they might want a fresh identity?
			// "If all fields are ANY and Randomize=true -> generate completely new valid identity."
			// But this is usually for CREATION time.

			// If the user *changes* parameters, e.g. switches from Chrome to Safari,
			// we should probably update the session's identity or error?
			// "Session ... Reuse that Identity consistently"

			// For now, we return the existing session. Rotation is an explicit action we can add later
			// or handle if specific opts flag "ForceRotate".
			return sess, nil
		}
	}

	// 2. Create new session
	jar := cookies.NewSessionJar()

	// Generate Identity
	ident := identity.Generate(config)

	// Create Transport with Identity's TLS Fingerprint
	transport := client.NewTransport(ident.ClientHelloID)

	httpClient := &http.Client{
		Transport: transport,
		Jar:       jar,
		Timeout:   timeout,
	}

	sess := &Session{
		ID:       id,
		Client:   httpClient,
		Jar:      jar,
		Identity: ident,
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
	sess, err := getOrCreateSession(opts.SessionID, opts.Identity, opts.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to init session: %w", err)
	}

	// Debug Hook
	if opts.Debug && opts.Hooks.OnRequest != nil {
		// We can't easily hook *before* Request creation effectively for logging everything
		// unless we create a dummy request.
	}

	// 2. Prepare Request
	req, err := http.NewRequest(opts.Method, urlStr, bytes.NewBuffer(opts.Body))
	if err != nil {
		return nil, err
	}

	// 3. Generate Headers (Using Session Identity)
	// We do NOT regenerate headers randomly per request if we want consistency.
	// We generate them based on the Identity.
	browserHeaders := headers.Generate(sess.Identity)
	for k, v := range browserHeaders {
		req.Header[k] = v
	}

	// Apply user overrides
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	// 4. Proxy Selection
	currentProxy := opts.ProxyURL
	proxyList := opts.Proxies

	rotateProxy := func() {
		if len(proxyList) > 0 {
			currentProxy = proxyList[0]
			if len(proxyList) > 1 {
				// Rotate list
				proxyList = append(proxyList[1:], proxyList[0])
			}
		}
	}

	// Ensure we start with a proxy if list provided but URL empty
	if currentProxy == "" && len(proxyList) > 0 {
		rotateProxy()
	}

	// 5. Execution Loop (Retry Logic)
	var httpResp *http.Response
	policy := retry.DefaultPolicy()

	for i := 0; i <= policy.MaxAttempts; i++ {
		// Hooks
		if opts.Hooks.OnRequest != nil {
			opts.Hooks.OnRequest(req)
		}

		// Context with Proxy
		ctx := req.Context()
		if currentProxy != "" {
			ctx = context.WithValue(ctx, client.CtxProxyURL, currentProxy)
		}
		reqWithCtx := req.WithContext(ctx)

		httpResp, err = sess.Client.Do(reqWithCtx)

		// Handle Network Errors
		if err != nil {
			if opts.Debug {
				fmt.Printf("[Debug] Attempt %d failed: %v\n", i+1, err)
			}
			if opts.Hooks.OnRetry != nil {
				opts.Hooks.OnRetry(i+1, err)
			}

			if i < policy.MaxAttempts {
				rotateProxy() // Network error usually means bad proxy?
				continue
			}
			return nil, err
		}

		// Handle Blocking / Retry Codes
		if policy.ShouldRetry(httpResp.StatusCode) {
			httpResp.Body.Close()

			if opts.Debug {
				fmt.Printf("[Debug] Attempt %d blocked: %d\n", i+1, httpResp.StatusCode)
			}
			if opts.Hooks.OnRetry != nil {
				opts.Hooks.OnRetry(i+1, fmt.Errorf("status %d", httpResp.StatusCode))
			}

			if i < policy.MaxAttempts {
				rotateProxy()
				// TODO: Implement Identity Rotation here if configured?
				// softRotate vs hardRotate?
				time.Sleep(1 * time.Second)
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

	if opts.Hooks.OnResponse != nil {
		opts.Hooks.OnResponse(httpResp)
	}

	// Debug Info
	if opts.Debug {
		fmt.Printf("[Debug] Final Status: %d\n", httpResp.StatusCode)
		fmt.Printf("[Debug] Identity: %s on %s (%s)\n", sess.Identity.Browser, sess.Identity.OS, sess.Identity.Device)
		fmt.Printf("[Debug] Proxy: %s\n", currentProxy)
	}

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
