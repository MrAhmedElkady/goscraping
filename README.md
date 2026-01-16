# goscraping

A Go library for high-scale scraping, designed as a near drop-in replacement for Node.js `got-scraping`.

## Features

- **Browser-like TLS Fingerprints**: Mimics Chrome 120 and Safari 16 using `uTLS`.
- **Automatic Header Generation**: Generates realistic browser headers (UA, Sec-Ch-Ua, etc.) based on profiles.
- **Session Support**: Maintains separate cookies and connection pools per session ID.
- **Proxy Support**: HTTP, HTTPS, and SOCKS5 proxy support with automatic "CONNECT" tunneling for uTLS.
- **Built-in Retries**: customizable retry policy for 429, 503, etc.

## Installation

```bash
go get goscraping
```

## Usage

```go
package main

import (
	"fmt"
	"time"
	"goscraping"
)

func main() {
	resp, err := goscraping.Fetch("https://example.com", &goscraping.Options{
		Method:        "GET",
		SessionID:     "my-session",
		HeaderProfile: "chrome", // or "safari"
		Timeout:       30 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Println(string(resp.Body))
}
```

## Options

- `Method`: HTTP method (GET, POST, etc.)
- `Headers`: Custom headers to override or append to browser defaults.
- `Body`: Request body (byte slice).
- `ProxyURL`: URL of the proxy (http, https, socks5).
- `SessionID`: String ID to persist cookies and connection state.
- `HeaderProfile`: "chrome" or "safari".

## Disclaimer

This library approximates browser fingerprints. For 100% undetectability against advanced bot protection, integration with a headless browser might be required, but `goscraping` aims to solve 99% of use cases with much lower resource usage.
