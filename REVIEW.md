# Project Review - goscraping

## Architectural Overview

The library has been refactored to separate concerns cleanly:

- **Client**: Handles the low-level transport, `uTLS` integration, and proxy dialing instructions. It exposes a `Transport` that can be configured with a Context-aware proxy dialer.
- **Headers**: A dynamic generation system (`Generate`) that constructs realistic browser headers based on a configuration (Browser/OS/Device).
- **Session**: Manages Identity (Cookies, Fingerprint) vs Network (Transport, Proxy). It uses a lock-protected store for thread safety.
- **Fetch**: The high-level orchestrator that combines Session, Headers, and Retry logic.

## Improvements Made

1.  **Dynamic Headers**: Replaced static profiles with a flexible generator supporting Chrome/Safari on various OSs.
2.  **Context-Aware Proxying**: The `Transport` can now route requests through different proxies (per request) via `Context`, allowing for rotation without destroying the session/connection pool.
3.  **Stable Fingerprints**: Sessions now persist their assigned TLS fingerprint (simulated) to avoid detection caused by changing fingerprints mid-session.
4.  **Robust Retries**: Implemented retry logic for 429/503/403 with proxy rotation hooks.

## Key Design Decisions

- **No Headless**: Strictly adhered to HTTP-only for performance.
- **uTLS**: Chosen for TLS fingerprinting. Note that `uTLS` has some limitations with generic `http.Transport` (HTTP/2), which we worked around by manually dialing and handing the connection to `RoundTrip` (via `DialTLSContext`).
- **Context for Proxies**: To support rotation while keeping `Keep-Alive` connections (where possible) and Session reuse, proxies are passed via Context.

## Known Limitations

- **HTTP/2 Ordering**: While we enabled HTTP/2, strict header frame ordering matching Chrome exactly requires lower-level access than `net/http` typically affords without a fork. We rely on standard Go sorting which is "good enough" for many cases but not perfect.
- **Cookie Jar**: References `net/http/cookiejar` which is compliant but basic.

## Future Enhancements

- **Header Ordering**: Implement a custom `RoundTripper` that writes raw bytes to enforcing strict header order for HTTP/1.1.
- **Firefox Support**: Add Firefox ClientHello ID and headers.
- **JS Tokens**: If sites require JS-calculated tokens, this library cannot provide them. Integration with an external token solver would be needed.
