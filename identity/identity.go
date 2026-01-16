package identity

import (
	"goscraping/types"

	utls "github.com/refraction-networking/utls"
)

// Identity represents a simulated browser user with specific versions
// This represents ONE coherent, realistic browser instance
type Identity struct {
	// High-level characteristics
	Browser types.Browser
	Device  types.Device
	OS      types.OS
	Locale  string

	// Version information (NEW - critical for realism)
	BrowserVersion types.BrowserVersion
	OSVersion      types.OSVersion

	// Generated values (derived from above)
	// FingerprintHash is a short unique ID for this generated identity
	FingerprintHash string

	UserAgent       string
	SecChUa         string
	SecChUaMobile   string
	SecChUaPlatform string

	// TLS Fingerprint ID (must align with browser version)
	ClientHelloID utls.ClientHelloID
}
