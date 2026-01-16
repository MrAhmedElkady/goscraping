# goscraping

A production-grade Go scraping library designed as a near drop-in replacement for Node.js `got-scraping`.

## Features

- **Browser TLS Fingerprints**: Mimics Chrome 120+ and Safari 16+ using `uTLS`.
- **Dynamic Header Generation**: Generates realistic headers based on Browser, OS, Device, and Locale.
- **Session Management**: Persistent cookies and connection pooling with stable per-session fingerprints.
- **Proxy Support**: HTTP, HTTPS, and SOCKS5 proxies with rotation support.
- **Retry Mechanism**: Robust handling of 403, 429, and 503 errors.
- **HTTP/2 Support**: Tuned to approximate browser behavior.

## Installation

```bash
go get github.com/MrAhmedElkady/goscraping
```

## Comparisons

| Feature | got-scraping (Node.js) | goscraping (Go) |
|---------|------------------------|-----------------|
| Language | JavaScript/TypeScript | Go |
| TLS Fingerprints | Yes | Yes (uTLS) |
| Header generation | Yes | Yes (Dynamic) |
| Session reuse | Yes | Yes |
| JS Execution | No | No |
| Headless Browser | No | No |

## Usage

### Simple Fetch

```go
package main

import (
	"fmt"
	"goscraping"
    "goscraping/types"
)

func main() {
	resp, err := goscraping.Fetch("https://example.com", &goscraping.Options{
		HeaderConfig: types.HeaderConfig{
			Browser: "chrome",
			Device:  "desktop",
			OS:      "windows",
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.StatusCode)
}
```

### Proxy & Rotation

```go
resp, err := goscraping.Fetch("https://example.com", &goscraping.Options{
    ProxyURL: "http://user:pass@1.2.3.4:8080",
    Proxies: []string{
        "http://proxy1:8080",
        "http://proxy2:8080",
    },
    SessionID: "my-session",
})
```

## Disclaimer

This library mimics browser network signatures but does **not** execute JavaScript. It is designed for high-scale, efficient scraping of content that is accessible via HTTP.
