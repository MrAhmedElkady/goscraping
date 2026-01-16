package identity

import (
	"math/rand"

	"goscraping/types"
)

// Generate creates a new Identity based on config constraints
// This is the CORE of the fingerprinting system
func Generate(cfg types.IdentityConfig) Identity {
	// Initialize RNG if needed (using package-level rng from versions.go)

	// 1. Resolve Browser/Device/OS Constraints
	// If "Any", pick random but realistic combination
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
	// This is CRITICAL - we must never generate impossible combinations
	browser, os, device = correctInvalidCombinations(browser, os, device)

	// 3. Select Versions
	// This is NEW and crucial for realism
	browserVer := selectBrowserVersion(browser)
	osVer := selectOSVersion(os, device)

	// 4. Validate compatibility (e.g., Chrome 120 needs Android 10+)
	if !validateVersionCompatibility(browser, browserVer, os, osVer) {
		// If invalid, adjust OS version to be compatible
		osVer = selectOSVersion(os, device)
	}

	// 5. Build Identity struct
	id := Identity{
		Browser:        browser,
		Device:         device,
		OS:             os,
		Locale:         cfg.Locale,
		BrowserVersion: browserVer,
		OSVersion:      osVer,
	}
	if id.Locale == "" {
		id.Locale = "en-US,en;q=0.9"
	}

	// 6. Generate derived values
	// User-Agent must reflect versions
	id.UserAgent = generateUserAgent(browser, browserVer, os, osVer, device)

	// TLS Fingerprint must match browser version
	id.ClientHelloID = selectTLSFingerprint(browser, browserVer, os)

	// Populate Chrome-specific headers if applicable
	populateChromeHeaders(&id)

	return id
}

func pickRandomBrowser() types.Browser {
	// Weighted selection: Chrome is more common than Safari
	if rand.Intn(10) < 7 {
		return types.BrowserChrome
	}
	return types.BrowserSafari
}

func pickRandomDevice(b types.Browser) types.Device {
	// Weighted selection: Desktop slightly more common
	if rand.Intn(10) < 6 {
		return types.DeviceDesktop
	}
	return types.DeviceMobile
}

func pickRandomOS(b types.Browser, d types.Device) types.OS {
	if b == types.BrowserSafari {
		// Safari only on macOS or iOS
		if d == types.DeviceMobile || d == types.DeviceTablet {
			return types.OSiOS
		}
		return types.OSMacOS
	}

	// Chrome
	if d == types.DeviceMobile {
		return types.OSAndroid
	}

	// Desktop Chrome
	// Windows, macOS, or Linux
	r := rand.Intn(10)
	if r < 6 {
		return types.OSWindows
	} else if r < 9 {
		return types.OSMacOS
	}
	return types.OSLinux
}

// correctInvalidCombinations fixes impossible browser/OS/device combos
// This ensures we NEVER generate something like "Safari on Android"
func correctInvalidCombinations(browser types.Browser, os types.OS, device types.Device) (types.Browser, types.OS, types.Device) {
	// Rule 1: Safari only on macOS/iOS
	if browser == types.BrowserSafari {
		if os == types.OSAndroid || os == types.OSWindows || os == types.OSLinux {
			// Conflict: Safari requested but incompatible OS
			// Priority: Keep browser, fix OS
			if device == types.DeviceMobile {
				os = types.OSiOS
			} else {
				os = types.OSMacOS
			}
		}
	}

	// Rule 2: iOS only with Safari (Chrome on iOS exists but has different UA/behavior)
	// For now, we simplify: iOS = Safari
	if os == types.OSiOS {
		browser = types.BrowserSafari
		device = types.DeviceMobile
	}

	// Rule 3: Android is mobile-only
	if os == types.OSAndroid {
		device = types.DeviceMobile
	}

	return browser, os, device
}

// populateChromeHeaders fills Chrome-specific headers
func populateChromeHeaders(id *Identity) {
	if id.Browser != types.BrowserChrome {
		return
	}

	// Generate sec-ch-ua that matches browser version
	id.SecChUa = generateSecChUa(id.BrowserVersion)

	// Determine mobile flag
	if id.Device == types.DeviceMobile {
		id.SecChUaMobile = "?1"
	} else {
		id.SecChUaMobile = "?0"
	}

	// Platform header
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
