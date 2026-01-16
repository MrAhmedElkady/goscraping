package headers

import (
	"goscraping/identity"
	"goscraping/types"
	"net/http"
)

// Generate returns a set of headers based on the Identity
func Generate(id identity.Identity) http.Header {
	h := make(http.Header)

	// 1. Basic Headers
	h.Set("User-Agent", id.UserAgent)
	h.Set("Accept-Language", id.Locale)
	h.Set("Accept-Encoding", "gzip, deflate, br") // Standard

	// 2. Browser Specifics
	switch id.Browser {
	case types.BrowserChrome:
		h.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		h.Set("Sec-Ch-Ua", id.SecChUa)
		h.Set("Sec-Ch-Ua-Mobile", id.SecChUaMobile)
		h.Set("Sec-Ch-Ua-Platform", id.SecChUaPlatform)

		h.Set("Sec-Fetch-Site", "same-origin")
		h.Set("Sec-Fetch-Mode", "navigate")
		h.Set("Sec-Fetch-User", "?1")
		h.Set("Sec-Fetch-Dest", "document")
		h.Set("Upgrade-Insecure-Requests", "1")

	case types.BrowserSafari:
		h.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		h.Set("Sec-Fetch-Site", "same-origin")
		h.Set("Sec-Fetch-Dest", "document")
		h.Set("Sec-Fetch-Mode", "navigate")
	}

	return h
}
