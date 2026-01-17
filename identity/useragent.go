package identity

import (
	"fmt"

	"github.com/MrAhmedElkady/goscraping/types"
)

// generateUserAgent creates a realistic User-Agent string based on Identity
// This is where browser/OS version coupling happens
func generateUserAgent(browser types.Browser, browserVer types.BrowserVersion, os types.OS, osVer types.OSVersion, device types.Device) string {
	switch browser {
	case types.BrowserChrome:
		return generateChromeUA(browserVer, os, osVer, device)
	case types.BrowserSafari:
		return generateSafariUA(browserVer, os, osVer, device)
	default:
		// Fallback
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	}
}

// generateChromeUA generates Chrome User-Agent strings
// Chrome UA format varies by OS and device type
func generateChromeUA(browserVer types.BrowserVersion, os types.OS, osVer types.OSVersion, device types.Device) string {
	chromeVersion := fmt.Sprintf("%d.0.0.0", browserVer.Major)

	switch os {
	case types.OSAndroid:
		// Android Chrome Mobile UA Format:
		// Mozilla/5.0 (Linux; Android {VERSION}; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/{VERSION} Mobile Safari/537.36
		//
		// Note: "K" is Android's device model obfuscation (introduced in Android 10+)
		// Different Android versions may have subtle build differences
		return fmt.Sprintf("Mozilla/5.0 (Linux; Android %s; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Mobile Safari/537.36",
			osVer.Version, chromeVersion)

	case types.OSWindows:
		// Windows Chrome Desktop UA Format:
		// Mozilla/5.0 (Windows NT {VERSION}; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/{VERSION} Safari/537.36
		//
		// Windows 10 = NT 10.0
		// Windows 11 = NT 10.0 (Windows 11 doesn't change NT version in UA)
		return fmt.Sprintf("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36",
			chromeVersion)

	case types.OSMacOS:
		// macOS Chrome Desktop UA Format:
		// Mozilla/5.0 (Macintosh; Intel Mac OS X {VERSION}) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/{VERSION} Safari/537.36
		//
		// macOS version uses underscore format: 10_15_7, 11_0, 12_0, 13_0, 14_0
		return fmt.Sprintf("Mozilla/5.0 (Macintosh; Intel Mac OS X %s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36",
			osVer.Version, chromeVersion)

	case types.OSLinux:
		// Linux Chrome Desktop UA Format:
		// Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/{VERSION} Safari/537.36
		return fmt.Sprintf("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36",
			chromeVersion)

	default:
		return fmt.Sprintf("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36",
			chromeVersion)
	}
}

// generateSafariUA generates Safari User-Agent strings
// Safari is more limited: only macOS desktop and iOS mobile
func generateSafariUA(browserVer types.BrowserVersion, os types.OS, osVer types.OSVersion, device types.Device) string {
	safariVersion := fmt.Sprintf("%d.%d", browserVer.Major, browserVer.Minor)

	switch os {
	case types.OSiOS:
		// iOS Safari Mobile UA Format:
		// Mozilla/5.0 (iPhone; CPU iPhone OS {VERSION} like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/{VERSION} Mobile/15E148 Safari/604.1
		//
		// iOS version format: 16_0, 16_1, etc. (underscores in UA)
		iosVersionUA := osVer.Version
		// Convert dot to underscore for UA
		if len(iosVersionUA) > 0 {
			// "16.0" -> "16_0"
			iosVersionUA = iosVersionUA[:len(iosVersionUA)-2] + "_" + iosVersionUA[len(iosVersionUA)-1:]
		}

		return fmt.Sprintf("Mozilla/5.0 (iPhone; CPU iPhone OS %s like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/%s Mobile/15E148 Safari/604.1",
			iosVersionUA, safariVersion)

	case types.OSMacOS:
		// macOS Safari Desktop UA Format:
		// Mozilla/5.0 (Macintosh; Intel Mac OS X {VERSION}) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/{VERSION} Safari/605.1.15
		return fmt.Sprintf("Mozilla/5.0 (Macintosh; Intel Mac OS X %s) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/%s Safari/605.1.15",
			osVer.Version, safariVersion)

	default:
		// Fallback to macOS Safari
		return fmt.Sprintf("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/%s Safari/605.1.15",
			safariVersion)
	}
}

// generateSecChUa generates the sec-ch-ua header value for Chrome
// This must match the browser version
func generateSecChUa(browserVer types.BrowserVersion) string {
	// Chrome sec-ch-ua format:
	// "Not_A Brand";v="8", "Chromium";v="{MAJOR}", "Google Chrome";v="{MAJOR}"
	// The "Not_A Brand" grease value varies slightly but v="8" is common
	return fmt.Sprintf(`"Not_A Brand";v="8", "Chromium";v="%d", "Google Chrome";v="%d"`,
		browserVer.Major, browserVer.Major)
}
