package fetch

import (
	"net/http"
)

// Response wraps http.Response
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	// Original http response if needed for streams (but we read body eagerly per requirements "body: []byte")
	RawResponse *http.Response
}
