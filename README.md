# goscraping

A production-grade, version-aware HTTP scraping library for Go.  
Designed as a near drop-in replacement for Node.js `got-scraping`.

**Philosophy**: Realistic browser fingerprinting through controlled entropy, not random chaos.

## Why Version-Aware Fingerprinting Matters

Traditional scraping libraries treat "Chrome on Android" as a single static identity. This is unrealistic and detectable at scale.

**In the real world:**
- Not everyone uses Android 13. Some use Android 11, 12, or 14.
- Not everyone has Chrome 121. Many have 118, 119, or 120.
- Different Android versions produce different User-Agent strings.
- Browser versions affect TLS fingerprints and HTTP headers.

**goscraping models this diversity realistically:**
- Each session gets a specific browser version (e.g., Chrome 118)
- Each session gets a specific OS version (e.g., Android 12)
- User-Agent, sec-ch-ua, and TLS fingerprints align with those versions
- Sessions maintain version consistency (no suspicious version drift)

## Installation

```bash
go get github.com/MrAhmedElkady/goscraping
```

## Quick Start

### Basic Usage (Chrome on Android)

```go
package main

import (
    "fmt"
    "goscraping"
    "goscraping/types"
)

func main() {
    resp, err := goscraping.Fetch("https://example.com", &goscraping.Options{
        SessionID: "my-session",
        Identity: types.IdentityConfig{
            Browser: types.BrowserChrome,
            Device:  types.DeviceMobile,
            OS:      types.OSAndroid,
            // Android version (10-14) and Chrome version (114-121) 
            // will be randomly selected within realistic ranges
        },
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.StatusCode)
}
```

**Result**: Session gets a specific identity like:
- **Browser**: Chrome 118
- **OS**: Android 12
- **User-Agent**: `Mozilla/5.0 (Linux; Android 12; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Mobile Safari/537.36`
- **sec-ch-ua**: `"Not_A Brand";v="8", "Chromium";v="118", "Google Chrome";v="118"`
- **TLS Fingerprint**: Chrome 118-compatible ClientHello

## Understanding User-Agent Differences

### Android Versions

Android versions affect the UA string:

```
Android 10: Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 ...
Android 11: Mozilla/5.0 (Linux; Android 11; K) AppleWebKit/537.36 ...
Android 12: Mozilla/5.0 (Linux; Android 12; K) AppleWebKit/537.36 ...
Android 13: Mozilla/5.0 (Linux; Android 13; K) AppleWebKit/537.36 ...
```

**Key Points:**
- The "K" is Android's device model obfuscation (Android 10+)
- OS version appears directly in the UA
- This creates natural diversity across sessions

### iOS Versions

iOS Safari UAs include both iOS and Safari versions:

```
iOS 16.0: Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) ... Version/16.0 ...
iOS 16.3: Mozilla/5.0 (iPhone; CPU iPhone OS 16_3 like Mac OS X) ... Version/16.0 ...
```

**Key Points:**
- iOS version uses underscore format (16_0, not 16.0)
- Safari version couples with iOS version
- Apple's versioning is more constrained than Android

### Desktop Browsers

Desktop UAs vary by OS and browser version:

**Chrome on Windows:**
```
Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 ... Chrome/120.0.0.0 ...
```

**Chrome on macOS:**
```
Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 ... Chrome/120.0.0.0 ...
```

**Safari on macOS:**
```
Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 ... Version/16.1 Safari/605.1.15
```

## Advanced Features

### Debug Mode

See exactly what identity was generated:

```go
opts := &goscraping.Options{
    Debug: true,
    Identity: types.IdentityConfig{
        Browser: types.BrowserChrome,
        Device:  types.DeviceMobile,
    },
}
```

Output:
```
[Debug] Identity: chrome on android (mobile)
[Debug] Browser: Chrome 118
[Debug] OS: Android 12
[Debug] User-Agent: Mozilla/5.0 (Linux; Android 12; K) ...
```

### Proxy Rotation

```go
opts := &goscraping.Options{
    Proxies: []string{
        "http://proxy1:8080",
        "socks5://proxy2:1080",
    },
    // On 403/429, proxy rotates automatically
}
```

### Session Persistence

```go
// First request: generates identity
resp1, _ := goscraping.Fetch(url, &goscraping.Options{
    SessionID: "user-123",
})

// Second request: reuses same identity (same versions, TLS, etc.)
resp2, _ := goscraping.Fetch(url2, &goscraping.Options{
    SessionID: "user-123",
})
```

## What goscraping CAN Do

✅ Mimic realistic browser network signatures  
✅ Generate version-diverse but valid identities  
✅ Align User-Agent, headers, and TLS fingerprints  
✅ Maintain session consistency (same identity across requests)  
✅ Rotate proxies intelligently on blocks  
✅ Support Chrome (114-121) and Safari (15-16)  

## What goscraping CANNOT Do

❌ Execute JavaScript  
❌ Render HTML or CSS  
❌ Solve CAPTCHAs  
❌ Bypass rate limits magically  
❌ Automate browser interactions  

**goscraping is an HTTP fingerprinting library, not a browser.**

## Disclaimer

This library is for educational and research purposes.  
It simulates realistic browser network behavior but cannot bypass systems requiring JavaScript execution or CAPTCHA solving.  
Use responsibly and respect robots.txt and terms of service.
