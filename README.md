# goscraping

A production-grade, version-consistent, and robust HTTP scraping library for Go.

## Robustness & Reliability

goscraping is designed to solve the common pitfalls of using `uTLS` with `net/http`:

1.  **Transport Manager (Strict HTTP/1.1)**:
    -   We explicitly force `HTTP/1.1` by default to prevent "malformed HTTP response" errors.
    -   We achieve this by patching the `uTLS` ClientHello spec to only advertise `http/1.1`, ensuring the server never incorrectly negotiates HTTP/2.
    -   A persistent `http.Transport` is managed to ensure connection reuse (Keep-Alive) while handling proxies via `CONNECT` tunneling.

2.  **Version-Aware Identity**:
    -    identities are generated with **weighted realism** (e.g., favoring Android 13/14 over Android 10).
    -   All Identity signals (User-Agent, sec-ch-ua, TLS Fingerprint) are strictly correlated to avoid detection.
    -   Identities are immutable per session to prevent behavior shifts.

3.  **Protocol-Safe Retries**:
    -   **Network Errors** (Timeout, DNS) -> Retry + Rotate Proxy.
    -   **Protocol Errors** (Malformed Response, Unsolicited Bytes) -> **Stop Immediately**. Failing fast prevents detection from infinite retry loops on fatal errors.

## Installation

```bash
go get github.com/MrAhmedElkady/goscraping
```

## Quick Start (Robust Fetch)

```go
package main

import (
    "fmt"
    "goscraping"
    "goscraping/types"
)

func main() {
    resp, err := goscraping.Fetch("https://httpbin.org/get", &goscraping.Options{
        SessionID: "chrome-user",
        Debug: true, // See protocol trace
        Identity: types.IdentityConfig{
            Browser: types.BrowserChrome,
        },
    })
    
    // Output will show: [Debug] Transport Trace: NegotiatedProtocol: "http/1.1"
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.StatusCode)
}
```

## Why Force HTTP/1.1?

Standard Go `net/http` Transport is not fully compatible with `uTLS` when HTTP/2 is negotiated via ALPN, because `uTLS` hides the negotiation details from the standard transport. goscraping's approach of strictly enforcing HTTP/1.1 is the industry-standard workaround for combining uTLS fingerprinting with the stability of `net/http`.
