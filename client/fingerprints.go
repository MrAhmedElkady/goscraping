package client

import (
	tls "github.com/refraction-networking/utls"
)

// Common fingerprints
var (
	FingerprintChrome120 = tls.HelloChrome_120
	FingerprintSafari16  = tls.HelloSafari_16_0
)

// GetFingerprint returns a client hello ID by name, defaulting to Chrome 120
func GetFingerprint(name string) tls.ClientHelloID {
	switch name {
	case "chrome_120", "chrome":
		return tls.HelloChrome_120
	case "safari_16", "safari":
		return tls.HelloSafari_16_0
	default:
		return tls.HelloChrome_120
	}
}
