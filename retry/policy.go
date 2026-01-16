package retry

// Policy defines when to retry a request
type Policy struct {
	MaxAttempts int
	RetryOn     []int
}

// DefaultPolicy returns a reasonable default retry policy
func DefaultPolicy() *Policy {
	return &Policy{
		MaxAttempts: 3,
		RetryOn:     []int{408, 429, 500, 502, 503, 504},
	}
}

// ShouldRetry checks if the response status code warrants a retry
func (p *Policy) ShouldRetry(statusCode int) bool {
	for _, code := range p.RetryOn {
		if statusCode == code {
			return true
		}
	}
	return false
}
