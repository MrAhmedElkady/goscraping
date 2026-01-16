package identity

import (
	"goscraping/client"
	"goscraping/types"

	utls "github.com/refraction-networking/utls"
)

// selectTLSProfile chooses the best TLS family for the requested browser
func selectTLSProfile(browser types.Browser, os types.OS) types.TLSProfile {
	switch browser {
	case types.BrowserChrome:
		// We fundamentally only have a few uTLS profiles.
		// We should pick one, then let the Version selector pick a matching version.

		// Randomly pick between recent modern Chrome (120) and slightly older (106)
		// to add diversity?
		// Or default to 120 as it's most modern.
		return types.TLSProfileChrome120

	case types.BrowserSafari:
		// Always Safari 16 for now
		return types.TLSProfileSafari16

	default:
		return types.TLSProfileChrome120
	}
}

// getClientHelloID returns the actual uTLS ID for a profile
func getClientHelloID(profile types.TLSProfile) utls.ClientHelloID {
	switch profile {
	case types.TLSProfileChrome120:
		return client.FingerprintChrome120
	case types.TLSProfileChrome106:
		return utls.HelloChrome_106_Shuffle
	case types.TLSProfileSafari16:
		return client.FingerprintSafari16
	default:
		return client.FingerprintChrome120
	}
}
