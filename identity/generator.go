package identity

import (
	"fmt"
	"math/rand"

	"github.com/MrAhmedElkady/goscraping/types"
)

// Generate creates a new Identity based on config constraints
func Generate(cfg types.IdentityConfig) Identity {
	// 1. Resolve Browser/Device/OS Constraints
	browser := cfg.Browser
	if browser == types.BrowserAny || browser == "" {
		browser = pickRandomBrowser()
	}

	device := cfg.Device
	if device == types.DeviceAny || device == "" {
		device = pickRandomDevice(browser)
	}

	os := cfg.OS
	if os == types.OSAny || os == "" {
		os = pickRandomOS(browser, device)
	}

	// 2. Correct Invalid Combinations
	browser, os, device = correctInvalidCombinations(browser, os, device)

	// 3. Select TLS Profile (The "Physical" Browser)
	// This is the source of truth for versions.
	tlsProfile := selectTLSProfile(browser, os)

	// 4. Select Versions COMPATIBLE with TLS Profile
	// This ensures no contradictions.
	browserVer := selectBrowserVersionForProfile(tlsProfile)

	osVer := selectOSVersion(os, device)

	// 5. Build Identity struct
	id := Identity{
		Browser:        browser,
		Device:         device,
		OS:             os,
		Locale:         cfg.Locale,
		BrowserVersion: browserVer,
		OSVersion:      osVer,
		ClientHelloID:  getClientHelloID(tlsProfile),
	}

	// Generate FingerprintHash
	id.FingerprintHash = fmt.Sprintf("%s_%s_%d.%d_%s",
		browser, os, browserVer.Major, browserVer.Minor, osVer.Version)

	if id.Locale == "" {
		id.Locale = "en-US,en;q=0.9"
	}

	// 6. Generate derived values
	id.UserAgent = generateUserAgent(browser, browserVer, os, osVer, device)
	populateChromeHeaders(&id)

	return id
}

func pickRandomBrowser() types.Browser {
	if rand.Intn(10) < 7 {
		return types.BrowserChrome
	}
	return types.BrowserSafari
}

func pickRandomDevice(b types.Browser) types.Device {
	if rand.Intn(10) < 6 {
		return types.DeviceDesktop
	}
	return types.DeviceMobile
}

func pickRandomOS(b types.Browser, d types.Device) types.OS {
	if b == types.BrowserSafari {
		if d == types.DeviceMobile || d == types.DeviceTablet {
			return types.OSiOS
		}
		return types.OSMacOS
	}

	if d == types.DeviceMobile {
		return types.OSAndroid
	}

	r := rand.Intn(10)
	if r < 6 {
		return types.OSWindows
	} else if r < 9 {
		return types.OSMacOS
	}
	return types.OSLinux
}

func correctInvalidCombinations(browser types.Browser, os types.OS, device types.Device) (types.Browser, types.OS, types.Device) {
	if browser == types.BrowserSafari {
		if os != types.OSMacOS && os != types.OSiOS {
			if device == types.DeviceMobile {
				os = types.OSiOS
			} else {
				os = types.OSMacOS
			}
		}
	}

	if os == types.OSiOS {
		browser = types.BrowserSafari
		device = types.DeviceMobile
	}

	if os == types.OSAndroid {
		device = types.DeviceMobile
	}

	return browser, os, device
}

func populateChromeHeaders(id *Identity) {
	if id.Browser != types.BrowserChrome {
		return
	}

	id.SecChUa = generateSecChUa(id.BrowserVersion)

	if id.Device == types.DeviceMobile {
		id.SecChUaMobile = "?1"
	} else {
		id.SecChUaMobile = "?0"
	}

	switch id.OS {
	case types.OSWindows:
		id.SecChUaPlatform = `"Windows"`
	case types.OSMacOS:
		id.SecChUaPlatform = `"macOS"`
	case types.OSAndroid:
		id.SecChUaPlatform = `"Android"`
	case types.OSLinux:
		id.SecChUaPlatform = `"Linux"`
	default:
		id.SecChUaPlatform = `"Unknown"`
	}
}
