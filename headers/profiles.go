package headers

import "net/http"

// Profile represents a set of browser headers
type Profile struct {
	UserAgent       string
	Platform        string
	SecChUa         string
	SecChUaMobile   string
	SecChUaPlatform string
	Accept          string
	AcceptLanguage  string
}

var (
	ChromeDesktop = Profile{
		UserAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		Platform:        `"Windows"`,
		SecChUa:         `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`,
		SecChUaMobile:   "?0",
		SecChUaPlatform: `"Windows"`,
		Accept:          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		AcceptLanguage:  "en-US,en;q=0.9",
	}

	SafariIOS = Profile{
		UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
		Platform:  "iPhone",
		// Safari often doesn't use sec-ch-ua headers as extensively as Chrome yet, or different ones.
		Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		AcceptLanguage: "en-US,en;q=0.9",
	}
)

// Apply applies the profile to the given http.Header
func (p Profile) Apply(h http.Header) {
	h.Set("User-Agent", p.UserAgent)
	h.Set("Accept", p.Accept)
	h.Set("Accept-Language", p.AcceptLanguage)

	// Sec headers
	if p.SecChUa != "" {
		h.Set("Sec-Ch-Ua", p.SecChUa)
	}
	if p.SecChUaMobile != "" {
		h.Set("Sec-Ch-Ua-Mobile", p.SecChUaMobile)
	}
	if p.SecChUaPlatform != "" {
		h.Set("Sec-Ch-Ua-Platform", p.SecChUaPlatform)
	}

	// Common fetch metadata (defaults)
	h.Set("Sec-Fetch-Site", "same-origin")
	h.Set("Sec-Fetch-Mode", "navigate")
	h.Set("Sec-Fetch-User", "?1")
	h.Set("Sec-Fetch-Dest", "document")
}
