package identity

import (
	"goscraping/client"
	"goscraping/types"

	utls "github.com/refraction-networking/utls"
)

// selectTLSFingerprint maps browser version to appropriate uTLS ClientHello ID
// This is critical for realistic fingerprinting
func selectTLSFingerprint(browser types.Browser, browserVer types.BrowserVersion, os types.OS) utls.ClientHelloID {
	switch browser {
	case types.BrowserChrome:
		// Chrome versions 114-121
		//
		// uTLS provides various Chrome fingerprints:
		// - HelloChrome_Auto (latest)
		// - HelloChrome_120
		// - HelloChrome_106_Shuffle
		// etc.
		//
		// For realism, we map version ranges to appropriate fingerprints
		// Note: Exact 1:1 mapping may not exist for all versions
		// We use the closest available fingerprint

		if browserVer.Major >= 120 {
			return client.FingerprintChrome120
		} else if browserVer.Major >= 106 {
			// Use Chrome 106 fingerprint for versions 106-119
			// This is an approximation but reasonable
			return utls.HelloChrome_106_Shuffle
		} else {
			// Fallback to a generic Chrome fingerprint
			return client.FingerprintChrome120
		}

	case types.BrowserSafari:
		// Safari fingerprints vary by OS/version
		// Safari on iOS uses different TLS than Safari on macOS

		if os == types.OSiOS {
			// iOS Safari
			// uTLS provides:
			// - HelloIOS_14 (older)
			// - HelloSafari_16_0 (macOS)
			//
			// For iOS 15-16, we use the Safari 16 fingerprint as closest match
			return client.FingerprintSafari16
		} else {
			// macOS Safari
			return client.FingerprintSafari16
		}

	default:
		// Fallback
		return client.FingerprintChrome120
	}
}

// IMPORTANT NOTE about TLS Fingerprint Approximations:
//
// uTLS ClientHello IDs are specific snapshots of browser versions.
// We cannot have a perfect 1:1 mapping for every single Chrome/Safari version.
//
// Our strategy:
// 1. Use the closest available fingerprint for the version range
// 2. Ensure consistency: same browser version = same fingerprint
// 3. Document approximations
//
// For example:
// - Chrome 114-119 -> Use Chrome 106 or 120 fingerprint (closest available)
// - Safari 15 -> Use Safari 16 fingerprint (close enough, same TLS stack)
//
// This is acceptable because:
// - Real users don't update browsers instantly
// - TLS stacks don't change with every minor version
// - Minor version differences in TLS are hard to detect
// - Consistency within a session is more important than exact matching
