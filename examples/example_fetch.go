package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/MrAhmedElkady/goscraping"
	"github.com/MrAhmedElkady/goscraping/types"
)

func main() {
	fmt.Println("=== goscraping Version-Aware Identity Demo ===")

	// Example 1: Chrome on Android (version will vary)
	fmt.Println("1. Chrome Mobile (Android with random version):")
	resp1, err := goscraping.Fetch("https://tls.peet.ws/api/all", &goscraping.Options{
		SessionID: "android-session-1",
		Identity: types.IdentityConfig{
			Browser: types.BrowserChrome,
			Device:  types.DeviceMobile,
			OS:      types.OSAndroid,
			// Browser and OS versions will be randomly selected within realistic ranges
		},
		Debug: true,
		Hooks: types.Hooks{
			OnRequest: func(req *http.Request) {
				fmt.Println("  User-Agent:", req.Header.Get("User-Agent"))
				fmt.Println("  sec-ch-ua:", req.Header.Get("Sec-Ch-Ua"))
			},
		},
	})
	if err != nil {
		fmt.Println("  Error:", err)
	} else {
		fmt.Printf("  Status: %d\n", resp1.StatusCode)
	}
	fmt.Println()

	// Example 2: Different Android session (will get different versions)
	fmt.Println("2. Another Chrome Mobile session (different Android version):")
	resp2, err := goscraping.Fetch("https://httpbin.org/headers", &goscraping.Options{
		SessionID: "android-session-2",
		Identity: types.IdentityConfig{
			Browser: types.BrowserChrome,
			Device:  types.DeviceMobile,
			OS:      types.OSAndroid,
		},
		Debug: true,
	})
	if err != nil {
		fmt.Println("  Error:", err)
	} else {
		fmt.Printf("  Status: %d\n\n", resp2.StatusCode)
	}

	// Example 3: Safari iOS
	fmt.Println("3. Safari on iOS:")
	resp3, err := goscraping.Fetch("https://httpbin.org/headers", &goscraping.Options{
		SessionID: "ios-session",
		Identity: types.IdentityConfig{
			Browser: types.BrowserSafari,
			Device:  types.DeviceMobile,
			OS:      types.OSiOS,
		},
		Debug: true,
	})
	if err != nil {
		fmt.Println("  Error:", err)
	} else {
		fmt.Printf("  Status: %d\n\n", resp3.StatusCode)
	}

	// Example 4: Desktop Chrome on Windows
	fmt.Println("4. Chrome Desktop (Windows):")
	resp4, err := goscraping.Fetch("https://httpbin.org/headers", &goscraping.Options{
		SessionID: "desktop-session",
		Identity: types.IdentityConfig{
			Browser: types.BrowserChrome,
			Device:  types.DeviceDesktop,
			OS:      types.OSWindows,
		},
		Debug:   true,
		Timeout: 20 * time.Second,
	})
	if err != nil {
		fmt.Println("  Error:", err)
	} else {
		fmt.Printf("  Status: %d\n", resp4.StatusCode)
	}

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("Notice how each session gets:")
	fmt.Println("- Different Android/iOS/browser versions")
	fmt.Println("- Consistent User-Agent and sec-ch-ua values")
	fmt.Println("- TLS fingerprints matching the browser version")
}
