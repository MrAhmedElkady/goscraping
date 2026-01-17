package identity

import (
	"math/rand"
	"time"

	"github.com/MrAhmedElkady/goscraping/types"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// selectBrowserVersionForProfile picks a browser version COMPATIBLE with the chosen TLS profile
func selectBrowserVersionForProfile(profile types.TLSProfile) types.BrowserVersion {
	var vRange types.VersionRange

	switch profile {
	case types.TLSProfileChrome120:
		vRange = types.Chrome120Range
		// Weighted: 70% chance of latest (MaxMajor), 30% others
		r := rng.Float64()
		major := vRange.MaxMajor
		if r < 0.3 {
			diff := vRange.MaxMajor - vRange.MinMajor
			if diff > 0 {
				major = vRange.MinMajor + rng.Intn(diff)
			}
		}
		return types.BrowserVersion{Major: major, Minor: 0, Patch: 0}

	case types.TLSProfileChrome106:
		vRange = types.Chrome106Range
		major := rng.Intn(vRange.MaxMajor-vRange.MinMajor+1) + vRange.MinMajor
		return types.BrowserVersion{Major: major, Minor: 0, Patch: 0}

	case types.TLSProfileSafari16:
		vRange = types.Safari16Range
		// Safari 16.x preferred (80%), Safari 15.x (20%)
		r := rng.Float64()
		major := 16
		if r < 0.2 {
			major = 15
		}

		minor := 0
		if major == 16 {
			// Safari 16.6 is very common, or random newer 16.x
			if rng.Float64() < 0.7 {
				minor = 6 // 16.6
			} else {
				minor = rng.Intn(4 + 1) // 0..4
			}
		} else {
			// Safari 15.x -> 15.4, 15.5
			minor = rng.Intn(2) + 4
		}
		return types.BrowserVersion{Major: major, Minor: minor, Patch: 0}

	default:
		// Fallback safe defaults
		return types.BrowserVersion{Major: 120, Minor: 0, Patch: 0}
	}
}

// selectOSVersion picks a realistic OS version
func selectOSVersion(os types.OS, device types.Device) types.OSVersion {
	switch os {
	case types.OSAndroid:
		// Favor Android 13/14 (indices 3,4) -> 60%
		r := rng.Float64()
		idx := 0
		if r < 0.6 && len(types.AndroidVersions) >= 5 {
			idx = 3 + rng.Intn(2)
		} else {
			idx = rng.Intn(len(types.AndroidVersions))
		}
		return types.OSVersion{Version: types.AndroidVersions[idx]}

	case types.OSiOS:
		// Favor newest versions (last 3) -> 70%
		n := len(types.IOSVersions)
		r := rng.Float64()
		idx := 0
		if r < 0.7 && n >= 3 {
			idx = n - 1 - rng.Intn(3)
		} else {
			idx = rng.Intn(n)
		}
		return types.OSVersion{Version: types.IOSVersions[idx]}

	case types.OSWindows:
		idx := rng.Intn(len(types.WindowsVersions))
		return types.OSVersion{Version: types.WindowsVersions[idx]}

	case types.OSMacOS:
		idx := rng.Intn(len(types.MacOSVersions))
		return types.OSVersion{Version: types.MacOSVersions[idx]}

	case types.OSLinux:
		return types.OSVersion{Version: ""}

	default:
		return types.OSVersion{Version: "10"}
	}
}
