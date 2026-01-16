package types

import (
	"net/http"
	"time"
)

// Hooks defines callbacks for request lifecycle events
type Hooks struct {
	OnRequest  func(*http.Request)
	OnResponse func(*http.Response)
	OnRetry    func(attempt int, err error)
}

// Options configuration for the Fetch request
type Options struct {
	// Methods & URL
	Method string
	Body   []byte

	// Identity Configuration
	// Who are we pretending to be?
	Identity  IdentityConfig
	SessionID string // Persistent session ID for cookies/identity reuse

	// Network Configuration
	Timeout  time.Duration
	ProxyURL string   // Single proxy to use
	Proxies  []string // List of proxies to rotate

	// Retry Configuration
	// TODO: Move MaxAttempts here if we want per-request override
	// Currently it's in retry.Policy defaults.

	// Headers
	Headers map[string]string

	// Debugging
	Debug bool
	Hooks Hooks
}

// DefaultOptions returns safe defaults
func DefaultOptions() *Options {
	return &Options{
		Method:  "GET",
		Timeout: 30 * time.Second,
		Identity: IdentityConfig{
			Browser: BrowserAny,
			Device:  DeviceAny,
			OS:      OSAny,
		},
		Headers: make(map[string]string),
	}
}

// IdentityConfig defines the constraints for generating an identity
// All fields are optional. Empty fields = Random selection.
type IdentityConfig struct {
	Browser Browser
	Device  Device
	OS      OS
	Locale  string
}
