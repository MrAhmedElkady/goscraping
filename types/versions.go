package types

import "fmt"

// BrowserVersion represents a browser version
type BrowserVersion struct {
	Major int
	Minor int
	Patch int
}

func (v BrowserVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// OSVersion represents an operating system version
type OSVersion struct {
	Version string
}

// VersionRange defines acceptable version ranges
type VersionRange struct {
	MinMajor int
	MaxMajor int
}

// TLSProfile represents a browser's TLS stack family
type TLSProfile string

const (
	TLSProfileChrome120 TLSProfile = "chrome_120"
	TLSProfileChrome106 TLSProfile = "chrome_106"
	TLSProfileSafari16  TLSProfile = "safari_16"
)

// Compatibility Maps
// These define SAFE, REALISTIC version ranges for a given TLS profile.
// DO NOT use a version outside these ranges for the corresponding profile.
var (
	// Chrome 120 TLS is compatible with versions ~117 to 123
	// We restrict to 118-121 to be safe and realistic
	Chrome120Range = VersionRange{MinMajor: 118, MaxMajor: 121}

	// Chrome 106 TLS (older shuffle) compatible with ~100 to 117
	Chrome106Range = VersionRange{MinMajor: 106, MaxMajor: 117}

	// Safari 16 TLS compatible with Safari 15.x and 16.x
	Safari16Range = VersionRange{MinMajor: 15, MaxMajor: 16}
)

// Android version compatibility
var AndroidVersions = []string{"10", "11", "12", "13", "14"}

// iOS version compatibility (major.minor format)
var IOSVersions = []string{"15.0", "15.6", "16.0", "16.1", "16.3", "16.6", "17.0"}

// Windows versions
var WindowsVersions = []string{"10", "11"}

// macOS versions (underscore format for UA)
var MacOSVersions = []string{"10_15_7", "11_0", "12_0", "13_0", "14_0"}
