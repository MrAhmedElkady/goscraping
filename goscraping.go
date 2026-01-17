package goscraping

import (
	"github.com/MrAhmedElkady/goscraping/fetch"
	"github.com/MrAhmedElkady/goscraping/types"
)

// Alias types for convenient top-level usage
type Options = types.Options
type Response = fetch.Response

// Forward Fetch to the internal implementation
func Fetch(url string, options *Options) (*Response, error) {
	return fetch.Fetch(url, options)
}
