package fetch

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/MrAhmedElkady/goscraping/client"
	"github.com/MrAhmedElkady/goscraping/cookies"
	"github.com/MrAhmedElkady/goscraping/headers"
	"github.com/MrAhmedElkady/goscraping/identity"
	"github.com/MrAhmedElkady/goscraping/retry"
	"github.com/MrAhmedElkady/goscraping/types"
	"github.com/andybalholm/brotli"
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
	transport := client.NewTransportManager(ident.ClientHelloID)

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
	req, err := http.NewRequest(opts.Method, urlStr, bytes.NewReader(opts.Body))
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

		// Reset Body for retries (Critical for Post requests)
		if req.GetBody != nil {
			if bodyC, err := req.GetBody(); err == nil {
				reqWithCtx.Body = bodyC
			}
		}

		httpResp, err = sess.Client.Do(reqWithCtx)

		// Determine Action
		action := policy.Classify(err, httpResp)

		switch action {
		case retry.Stop:
			if err != nil {
				// Fatal Error
				if opts.Debug && retry.IsProtocolError(err) {
					fmt.Printf("[Debug] Fatal Protocol Error: %v\n", err)
				}
				return nil, err
			}
			// Success
			goto Success

		case retry.Retry, retry.RotateAndRetry:
			// Log Failure
			if opts.Debug {
				reason := "Network/Status"
				if err != nil {
					reason = err.Error()
				} else if httpResp != nil {
					reason = fmt.Sprintf("Status %d", httpResp.StatusCode)
				}
				fmt.Printf("[Debug] Attempt %d failed (%s). Action: %v\n", i+1, reason, action)
			}

			if opts.Hooks.OnRetry != nil {
				opts.Hooks.OnRetry(i+1, err)
			}

			// Check if we exhausted retries
			if i == policy.MaxAttempts {
				if err != nil {
					return nil, err
				}
				return nil, fmt.Errorf("gave up after %d attempts, last status: %d", i+1, httpResp.StatusCode)
			}

			// Cleanup partial response
			if httpResp != nil {
				httpResp.Body.Close()
			}

			// Rotate Proxy if needed
			if action == retry.RotateAndRetry {
				rotateProxy()
			}

			// Backoff
			time.Sleep(time.Duration(i+1) * 500 * time.Millisecond)
			continue
		}
	}

Success:
	if err != nil {
		return nil, err
	}

	// Decompression Logic
	var reader io.ReadCloser
	switch httpResp.Header.Get("Content-Encoding") {
	case "gzip":
		gz, err := gzip.NewReader(httpResp.Body)
		if err == nil {
			reader = &readCloser{Reader: gz, Closer: httpResp.Body}
			// Update header to reflect decoded content
			httpResp.Header.Del("Content-Encoding")
			httpResp.Header.Del("Content-Length")
			httpResp.ContentLength = -1
			httpResp.Uncompressed = true
		} else {
			reader = httpResp.Body
		}
	case "deflate":
		fl := flate.NewReader(httpResp.Body)
		reader = &readCloser{Reader: fl, Closer: httpResp.Body}
		httpResp.Header.Del("Content-Encoding")
		httpResp.Header.Del("Content-Length")
		httpResp.ContentLength = -1
		httpResp.Uncompressed = true
	case "br":
		br := brotli.NewReader(httpResp.Body)
		reader = &readCloser{Reader: br, Closer: httpResp.Body}
		httpResp.Header.Del("Content-Encoding")
		httpResp.Header.Del("Content-Length")
		httpResp.ContentLength = -1
		httpResp.Uncompressed = true
	default:
		reader = httpResp.Body
	}
	httpResp.Body = reader
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

type readCloser struct {
	io.Reader
	io.Closer
}

func (rc *readCloser) Read(p []byte) (n int, err error) {
	return rc.Reader.Read(p)
}

func (rc *readCloser) Close() error {
	if c, ok := rc.Reader.(io.Closer); ok {
		c.Close()
	}
	return rc.Closer.Close()
}
