# REVIEW.md - Network Layer Refactor

## The "Malformed Response" Issue

**Problem**:
The library was previously creating `uTLS` connections that advertised `h2` (HTTP/2) support (because Chrome/Safari fingerprints include it). `net/http` was unaware of this negotiation and expected HTTP/1.1. When the server sent HTTP/2 frames (like SETTINGS), `net/http` read them as malformed HTTP/1.1 responses, closing the connection. The retry logic then naively retried indefinitely.

**Solution**:
1.  **Architecture**: Created a `Transport` which wraps a persistent `http.Transport`.
2.  **Spec Patching**: In `dialTLS`, we now retrieve the `ClientHelloSpec` from the fingerprint ID and **manually patch the ALPN extension** to remove `h2`.
3.  **Result**: `uTLS` now strictly offers `["http/1.1"]`. The server is forced to speak HTTP/1.1. `net/http` happy.

## Protocol-Aware Retries

We added `retry.IsProtocolError(err)`.
- If an error contains "malformed HTTP response" or "unsolicited response", we stop retrying immediately.
- This prevents spamming servers with broken handshakes.

## Proxy Tunneling

We implemented manual `HTTP CONNECT` logic in `client/transport.go` (via `proxy.go`).
- `http.Transport`'s built-in proxy support doesn't straightforwardly support custom `DialTLSContext` wrapping `uTLS`.
- We now control the full chain: `TCP Dial` -> `HTTP CONNECT` (if proxy) -> `uTLS Handshake` -> `Application Data`.

## Current Limitations

- **No HTTP/2**: Verification shows we use HTTP/1.1. This is arguably safer for scraping as H2 fingerpriting is complex. If H2 is strictly required, a dedicated H2 transport (using `golang.org/x/net/http2` directly) would be needed.
- **Spec Patching Overhead**: Parse/Patch of Spec adds negligible overhead but ensures correctness.
