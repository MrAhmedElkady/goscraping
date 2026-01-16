package types

import "time"

// OptionsConfig defines the configuration for a Fetch request
type Options struct {
	Method         string
	Headers        map[string]string
	Body           []byte
	ProxyURL       string
	SessionID      string
	Timeout        time.Duration
	FollowRedirect bool
	MaxRedirects   int

	// Explicit header ordering or profile selection could go here
	HeaderProfile string // e.g., "chrome", "safari"
}

// DefaultOptions returns default options
func DefaultOptions() *Options {
	return &Options{
		Method:         "GET",
		Timeout:        30 * time.Second,
		FollowRedirect: true,
		MaxRedirects:   10,
		HeaderProfile:  "chrome",
	}
}
