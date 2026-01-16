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
// Different OSes have different versioning schemes
type OSVersion struct {
	// For Android: Major version (10, 11, 12, 13)
	// For iOS: Major.Minor (16.0, 16.1, etc.)
	// For Windows: Major (10, 11)
	// For macOS: Major_Minor_Patch (10_15_7, etc.)
	Version string
}

// VersionRange defines acceptable version ranges
type VersionRange struct {
	MinMajor int
	MaxMajor int
}

// Common version ranges for browsers
var (
	// Chrome versions 114-121 (realistic range as of 2026)
	ChromeVersionRange = VersionRange{MinMajor: 114, MaxMajor: 121}

	// Safari versions 15-16
	SafariVersionRange = VersionRange{MinMajor: 15, MaxMajor: 16}
)

// Android version compatibility
var AndroidVersions = []string{"10", "11", "12", "13", "14"}

// iOS version compatibility (major.minor format)
var IOSVersions = []string{"15.0", "15.6", "16.0", "16.1", "16.3", "16.6", "17.0"}

// Windows versions
var WindowsVersions = []string{"10", "11"}

// macOS versions (underscore format for UA)
var MacOSVersions = []string{"10_15_7", "11_0", "12_0", "13_0", "14_0"}
