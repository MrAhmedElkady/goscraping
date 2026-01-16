package types

import (
	"time"
)

// OptionsConfig defines the configuration for a Fetch request
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

	// Fingerprint config
	HeaderConfig HeaderConfig // Use a custom struct or map
}

// HeaderConfig definitions matching headers package
type HeaderConfig struct {
	Browser string
	Device  string
	OS      string
	Locale  string
}

// DefaultOptions returns default options
func DefaultOptions() *Options {
	return &Options{
		Method:         "GET",
		Timeout:        30 * time.Second,
		FollowRedirect: true,
		MaxRedirects:   10,
		HeaderConfig: HeaderConfig{
			Browser: "chrome",
			Device:  "desktop",
			OS:      "windows",
		},
	}
}
