package retry

import (
	"net/http"
	"strings"
)

// Classification represents the type of action to take for an error
type Classification int

const (
	Retry          Classification = iota // Standard retry
	RotateAndRetry                       // Retry but rotate proxy first
	Stop                                 // Fatal error, stop immediately
)

// Policy defines when to retry a request
type Policy struct {
	MaxAttempts int
}

// DefaultPolicy returns a reasonable default retry policy
func DefaultPolicy() *Policy {
	return &Policy{
		MaxAttempts: 3,
	}
}

// Classify determines the action to take based on the error or status code
func (p *Policy) Classify(err error, resp *http.Response) Classification {
	// 1. Check for protocol errors (Fatal)
	if IsProtocolError(err) {
		return Stop
	}

	// 2. Check for Proxy errors (Rotate)
	if IsProxyError(err) {
		return RotateAndRetry
	}

	// 3. Check for Network errors (Standard Retry)
	// Most net errors like timeouts are simple retries,
	// but often rotating proxy helps if the proxy itself is slow.
	// For now, we'll treat generic network errors as RotateAndRetry
	// because usually it's the proxy's fault in scraping.
	if err != nil {
		return RotateAndRetry
	}

	// 4. Check Response Status Codes
	if resp != nil {
		switch resp.StatusCode {
		case 403, 407, 429, 500, 502, 503, 504:
			// 403/407/429 usually implies IP ban -> Rotate
			// 500s -> Server error, maybe Rotate helps if sticky session is bad
			return RotateAndRetry
		}
	}

	return Stop
}

// IsProtocolError returns true if the error indicates a fatal protocol mismatch
func IsProtocolError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())

	return strings.Contains(msg, "malformed http response") ||
		strings.Contains(msg, "unsolicited response") ||
		strings.Contains(msg, "too many redirects") ||
		strings.Contains(msg, "scheme") || // unsupported protocol scheme
		strings.Contains(msg, "tls: first record does not look like a tls handshake")
}

// IsProxyError checks for common proxy connection issues
func IsProxyError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())

	return strings.Contains(msg, "proxy connection failed") ||
		strings.Contains(msg, "proxyrefused") ||
		strings.Contains(msg, "socks connect") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "deadline exceeded") ||
		strings.Contains(msg, "connection reset")
}
