![goscraping](assets/main.png)

# 🕷️ goscraping

[![Go Report Card](https://goreportcard.com/badge/github.com/MrAhmedElkady/goscraping)](https://goreportcard.com/report/github.com/MrAhmedElkady/goscraping)
[![GoDoc](https://godoc.org/github.com/MrAhmedElkady/goscraping?status.svg)](https://godoc.org/github.com/MrAhmedElkady/goscraping)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> **Production-grade, stealthy HTTP scraping library for Go.**

`goscraping` is a high-performance scraping client designed to emulate real browser TLS fingerprints (Chrome, Safari, iOS, Android) using [uTLS](https://github.com/refraction-networking/utls). It solves the difficult problem of "TLS Fingerprinting" which often causes standard Go `http.Client` requests to be blocked by Cloudflare, Akamai, and other anti-bots.

---

## 📖 Table of Contents

- [Features](#-features)
- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [Advanced Usage](#-advanced-usage)
    - [Custom Identities](#custom-identities)
    - [Authenticated Proxies](#authenticated-proxies)
    - [Session Persistence](#session-persistence)
- [Architecture & Design](#-architecture--design)
    - [The TLS Fingerprint Problem](#the-tls-fingerprint-problem)
    - [Why Forced HTTP/1.1?](#why-forced-http11)
- [Best Practices](#-best-practices)
- [License](#-license)

---

## 🚀 Features

-   **🎭 Browser Impersonation**: Automatically mimics the **TLS Client Hello** packets of real browsers (Chrome 120+, Safari 17+, etc.).
-   **🔄 Smart Session Management**: Maintains a consistent "Identity" (User-Agent, Headers, TLSJA3) throughout a session.
-   **🛡️ Stealth Mode**: Strictly forces **HTTP/1.1** to prevent HTTP/2 fingerprint leaks (a common detection vector).
-   **⚡ Robust Fetching**:
    -   Exponential Backoff Retries.
    -   Auto-rotation of dead proxies.
    -   Fail-fast on fatal protocol errors.
-   **📦 Transparent Decompression**: seamless support for `gzip`, `deflate`, and `brotli` (br) encoding.
-   **🍪 Cookie Jar**: Built-in, persistent cookie handling per session.

---

## 📦 Installation

```bash
go get github.com/MrAhmedElkady/goscraping
```

---

## ⚡ Quick Start

```go
package main

import (
    "fmt"
    "time"

    "github.com/MrAhmedElkady/goscraping"
    "github.com/MrAhmedElkady/goscraping/types"
)

func main() {
    // 1. Configure Options
    opts := types.DefaultOptions()
    opts.Timeout = 15 * time.Second
    opts.Method = "GET"
    
    // Optional: Enable debug logs to see the magic
    opts.Debug = true

    // 2. Fetch URL
    resp, err := goscraping.Fetch("https://httpbin.org/get", opts)
    if err != nil {
        panic(err)
    }
    // resp.Body is fully read and closed, safe to use immediately.

    fmt.Printf("Status: %d\n", resp.StatusCode)
    fmt.Printf("Body: %s\n", string(resp.Body))
}
```

---

## 🛠 Advanced Usage

### Custom Identities

You can force the scraper to behave like a specific device (e.g., iPhone using Safari).

```go
opts.Identity = types.IdentityConfig{
    Browser: types.BrowserSafari,
    OS:      types.OSMacOS,
    Device:  types.DeviceDesktop,
}
```

### Authenticated Proxies

We support `http` and `socks5` proxies with authentication.

```go
opts.Proxies = []string{
    "http://user:pass@1.2.3.4:8080",
    "http://user:pass@5.6.7.8:8080",
}
// The library will automatically rotate through these if one fails.
```

### Session Persistence

To maintain cookies and identity (like a logged-in user) across multiple requests, use a `SessionID`.

```go
opts.SessionID = "user-session-123"

// Request 1: Login
goscraping.Fetch("https://example.com/login", opts)

// Request 2: Access Profile (Cookies & TLS fingerprint are preserved)
goscraping.Fetch("https://example.com/profile", opts)
```

---

## 🏗 Architecture & Design

### The TLS Fingerprint Problem
Standard Go tools (`net/http`) have a very distinct TLS handshake "fingerprint". Anti-bot systems see this and immediately know "This is a bot, not Chrome."

**goscraping** uses `uTLS` to byte-for-byte replica the handshake of real browsers. To the server, your request looks exactly like it came from Chrome 120 on Windows.

### Why Forced HTTP/1.1?
You might notice we forcibly downgrade to HTTP/1.1. 
**Why?**
When simulating a browser via `uTLS`, if we negotiate HTTP/2, we must also perfectly emulate the HTTP/2 frames (Window Update, Priority, Stream IDs). Go's standard `net/http2` does *not* perfectly match Chrome's HTTP/2 behavior. This discrepancy is a dead giveaway to anti-bots.

By forcing HTTP/1.1 (via ALPN patching), we force the server to talk the simpler protocol, closing this detection loophole completely.

---

## 🛡️ Best Practices

1.  **Use Sessions**: Don't create a new Identity for every request to the same site. It looks suspicious if the same IP changes from "Chrome on Windows" to "Safari on iPhone" instantly.
2.  **Wait**: Add random delays between requests.
3.  **Rotate Proxies**: If you get 403/407 errors, your IP is likely flagged. Use the `Proxies` list to auto-rotate.

---

## 🗺️ Roadmap & Status

This project is actively maintained. Below is the current feature status:

- [x] **Advanced TLS Fingerprinting**: Full uTLS integration for Chrome/Safari/iOS mimicry.
- [x] **Strict HTTP/1.1 Engine**: custom ALPN negotiation to prevent HTTP/2 fingerprint leaks.
- [x] **Authenticated Proxy Support**: full support for HTTP/SOCKS5 proxies with rotation.
- [x] **Smart Decompression**: transparent handling of `brotli`, `gzip`, and `deflate`.
- [x] **Production Stability**: fail-fast logic for protocol errors and robust retries.
- [ ] **Headless Browser**: future support for JavaScript rendering (via Chrome DevTools Protocol).
- [ ] **CAPTCHA Solving**: integrated API hooks for major solving services.

---

## 📄 License

MIT © 2026 Ahmed Elkady.
