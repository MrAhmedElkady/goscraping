package types

import (
	"net/http"
	"time"
)

// IdentityConfig defines preferences for generating a browser identity
type IdentityConfig struct {
	Browser Browser // chrome | safari | firefox | any
	Device  Device  // desktop | mobile | tablet | any
	OS      OS      // windows | macos | linux | android | ios | any
	Locale  string  // e.g., "en-US"

	Randomize bool   // Generate a new identity if one doesn't exist or forced
	Stable    bool   // Lock identity to session
	Seed      string // Optional seed for deterministic generation
}

// Hooks defines callbacks for observability
type Hooks struct {
	OnRequest  func(req *http.Request)
	OnResponse func(resp *http.Response)
	OnRetry    func(attempt int, err error)
}

// Options defines the configuration for a Fetch request
type Options struct {
	Method  string
	Headers map[string]string
	Body    []byte

	// Proxy & Rotation
	ProxyURL string
	Proxies  []string // List of proxies to rotate through

	// Session
	SessionID string

	// Behavior config
	Timeout        time.Duration
	FollowRedirect bool
	MaxRedirects   int

	// Identity / Fingerprint config
	Identity IdentityConfig

	// Observability
	Debug bool
	Hooks Hooks
}

// DefaultOptions returns default options
func DefaultOptions() *Options {
	return &Options{
		Method:         "GET",
		Timeout:        30 * time.Second,
		FollowRedirect: true,
		MaxRedirects:   10,
		Identity: IdentityConfig{
			Browser: BrowserChrome,
			Device:  DeviceDesktop,
			OS:      OSWindows,
		},
	}
}
