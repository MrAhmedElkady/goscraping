package headers

import (
	"math/rand"
	"net/http"
	"time"
)

// Config defines the parameters for header generation
type Config struct {
	Browser string
	OS      string
	Device  string
	Locale  string // e.g., "en-US"
}

// Generate returns a set of headers based on the config
func Generate(cfg Config) http.Header {
	h := make(http.Header)

	// Defaults
	if cfg.Browser == "" {
		cfg.Browser = Chrome
	}
	if cfg.OS == "" {
		cfg.OS = Windows
	}
	if cfg.Device == "" {
		cfg.Device = Desktop
	}
	if cfg.Locale == "" {
		cfg.Locale = DefaultAcceptLanguage
	}

	// Logic to select UA and Platform values
	var ua, secChUa, secChUaPlatform, secChUaMobile string

	// This simple logic can be expanded to a lookup table
	switch cfg.Browser {
	case Chrome:
		if cfg.Device == Desktop {
			if cfg.OS == MacOS {
				ua = UA_Chrome_MacOS
				secChUa = SecChUa_Chrome_MacOS
				secChUaPlatform = SecChUaPlatform_MacOS
			} else {
				ua = UA_Chrome_Windows
				secChUa = SecChUa_Chrome_Windows
				secChUaPlatform = SecChUaPlatform_Windows
			}
			secChUaMobile = "?0"
		} else {
			// Mobile
			ua = UA_Chrome_Android
			secChUa = SecChUa_Chrome_Mobile
			secChUaPlatform = SecChUaPlatform_Android
			secChUaMobile = "?1"
		}

	case Safari:
		if cfg.Device == Mobile || cfg.OS == iOS {
			ua = UA_Safari_iOS
		} else {
			ua = UA_Safari_MacOS
		}
		// Safari doesn't typically send Sec-Ch-Ua
	}

	// Common Headers
	// Note: Order in the map is random. Writing to the wire is done by the Transport.
	// If strict order is required, we need a custom Transport wrapper that writes these specific keys in order.
	// For now, we populate the map.

	h.Set("User-Agent", ua)
	h.Set("Accept-Language", cfg.Locale)
	h.Set("Accept-Encoding", DefaultAcceptEncoding)

	// Browser Specifics
	if cfg.Browser == Chrome {
		h.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		h.Set("Sec-Ch-Ua", secChUa)
		h.Set("Sec-Ch-Ua-Mobile", secChUaMobile)
		h.Set("Sec-Ch-Ua-Platform", secChUaPlatform)
		h.Set("Sec-Fetch-Site", "same-origin")
		h.Set("Sec-Fetch-Mode", "navigate")
		h.Set("Sec-Fetch-User", "?1")
		h.Set("Sec-Fetch-Dest", "document")
		// Upgrade-Insecure-Requests is common
		h.Set("Upgrade-Insecure-Requests", "1")
	} else if cfg.Browser == Safari {
		h.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		h.Set("Sec-Fetch-Site", "same-origin")
		// Safari varies on these
		h.Set("Sec-Fetch-Dest", "document")
		h.Set("Sec-Fetch-Mode", "navigate")
	}

	return h
}

// RandomBrowserConfig returns a random realistic configuration
func RandomBrowserConfig() Config {
	// Weighted random selection could go here.
	// Simple 50/50 for now
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if r.Intn(2) == 0 {
		return Config{Browser: Chrome, OS: Windows, Device: Desktop}
	}
	return Config{Browser: Safari, OS: MacOS, Device: Desktop}
}
