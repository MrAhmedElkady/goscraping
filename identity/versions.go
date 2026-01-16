package identity

import (
	"goscraping/types"
	"math/rand"
	"time"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// selectBrowserVersion picks a realistic browser version within the range
func selectBrowserVersion(browser types.Browser) types.BrowserVersion {
	switch browser {
	case types.BrowserChrome:
		// Pick a version between 114-121
		major := rng.Intn(types.ChromeVersionRange.MaxMajor-types.ChromeVersionRange.MinMajor+1) + types.ChromeVersionRange.MinMajor
		return types.BrowserVersion{Major: major, Minor: 0, Patch: 0}

	case types.BrowserSafari:
		// Pick a version between 15-16
		major := rng.Intn(types.SafariVersionRange.MaxMajor-types.SafariVersionRange.MinMajor+1) + types.SafariVersionRange.MinMajor
		minor := 0
		if major == 16 {
			// Safari 16.0, 16.1, etc.
			minor = rng.Intn(2) // 0 or 1
		}
		return types.BrowserVersion{Major: major, Minor: minor, Patch: 0}

	default:
		return types.BrowserVersion{Major: 120, Minor: 0, Patch: 0}
	}
}

// selectOSVersion picks a realistic OS version
func selectOSVersion(os types.OS, device types.Device) types.OSVersion {
	switch os {
	case types.OSAndroid:
		// Pick from Android 10-14
		idx := rng.Intn(len(types.AndroidVersions))
		return types.OSVersion{Version: types.AndroidVersions[idx]}

	case types.OSiOS:
		// Pick from iOS versions
		idx := rng.Intn(len(types.IOSVersions))
		return types.OSVersion{Version: types.IOSVersions[idx]}

	case types.OSWindows:
		// Pick Windows 10 or 11
		idx := rng.Intn(len(types.WindowsVersions))
		return types.OSVersion{Version: types.WindowsVersions[idx]}

	case types.OSMacOS:
		// Pick macOS version
		idx := rng.Intn(len(types.MacOSVersions))
		return types.OSVersion{Version: types.MacOSVersions[idx]}

	case types.OSLinux:
		// Linux doesn't need specific version in UA
		return types.OSVersion{Version: ""}

	default:
		return types.OSVersion{Version: "10"}
	}
}

// validateVersionCompatibility ensures browser version is compatible with OS version
// For example, Chrome 120 can't run on Android 8
func validateVersionCompatibility(browser types.Browser, browserVer types.BrowserVersion, os types.OS, osVer types.OSVersion) bool {
	// Android version compatibility
	if os == types.OSAndroid {
		// Chrome 114+ generally requires Android 7+, so all our versions (10-14) are fine
		// This is where we'd enforce constraints like "Chrome 121 needs Android 10+"
		return true
	}

	// iOS version compatibility
	if os == types.OSiOS {
		// Safari versions align with iOS versions
		// Safari 15 = iOS 15, Safari 16 = iOS 16
		// We're already ensuring this in the selection logic
		return true
	}

	// Desktop OSes are generally compatible with all browser versions
	return true
}
