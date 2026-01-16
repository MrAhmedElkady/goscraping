# goscraping

**goscraping** is a production-grade HTTP scraping library for Go, designed for
**protocol correctness, fingerprint consistency, and long-term stability at scale**.

Unlike many scraping tools that rely on brittle hacks, goscraping focuses on
**doing fewer things correctly**:
- Stable transports
- Realistic browser identities
- Explicit protocol control

---

## Core Principles

goscraping is built around three non-negotiable principles:

1. **Protocol Safety Over Feature Count**
2. **Statistical Realism Over Perfect Imitation**
3. **Explicit Design Over Implicit Magic**

If something cannot be implemented safely using `net/http` and `uTLS`,
it is documented — not hacked around.

---

## Robustness & Reliability

### 1. Strict Transport Management (HTTP/1.1 by Default)

goscraping explicitly enforces **HTTP/1.1** to avoid a class of fatal protocol errors
that occur when combining `uTLS` with Go’s standard `net/http` transport.

- The uTLS ClientHello is patched to advertise **only `http/1.1` via ALPN**
- This prevents servers from negotiating HTTP/2 unexpectedly
- A single persistent `http.Transport` is reused per session
- Proxy support is implemented via proper `CONNECT` tunneling (HTTP & SOCKS5)

> This design eliminates common errors such as:
> `malformed HTTP response`  
> `unsolicited response received on idle HTTP channel`

HTTP/2 support is intentionally **disabled by default** and may be introduced later
as a fully isolated transport.

---

### 2. Version-Aware Identity System

Each session is assigned a **stable, version-consistent browser identity**:

- Realistic, weighted version selection  
  (e.g. Android 13/14 > Android 10, recent Chrome > older releases)
- Full correlation between:
  - `User-Agent`
  - `sec-ch-ua`
  - OS / Device metadata
  - TLS ClientHello fingerprint
- Identity is **immutable for the lifetime of the session**

This prevents mid-session fingerprint drift, a common detection signal.

---

### 3. Protocol-Safe Retry Strategy

Retries are classified explicitly:

| Category | Behavior |
|--------|---------|
| Network errors (DNS, timeout) | Retry + optional proxy rotation |
| HTTP status (429, 503) | Retry with backoff |
| Proxy failures | Rotate proxy |
| **Protocol errors** | **Fail immediately (no retry)** |

Protocol errors include malformed responses, unsolicited bytes, or ALPN mismatches.
Failing fast is a deliberate design choice to avoid infinite retry loops that amplify
detection risk.

---

## Installation

```bash
go get github.com/MrAhmedElkady/goscraping
