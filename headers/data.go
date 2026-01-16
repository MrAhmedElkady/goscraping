package headers

// Browser constants
const (
	Chrome  = "chrome"
	Safari  = "safari"
	Firefox = "firefox"
)

// OS constants
const (
	Windows = "windows"
	MacOS   = "macos"
	Linux   = "linux"
	Android = "android"
	iOS     = "ios"
)

// Device constants
const (
	Desktop = "desktop"
	Mobile  = "mobile"
	Tablet  = "tablet"
)

// Data sets
var (
	// Simplified database of User-Agents and Brand strings.
	// In a full production lib, this might be loaded from a large JSON or embedded resource.

	// Chrome 120 (Desktop Windows)
	UA_Chrome_Windows       = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	SecChUa_Chrome_Windows  = `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`
	SecChUaPlatform_Windows = `"Windows"`

	// Chrome 120 (Desktop MacOS)
	UA_Chrome_MacOS       = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	SecChUa_Chrome_MacOS  = `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`
	SecChUaPlatform_MacOS = `"macOS"`

	// Safari 16 (Desktop MacOS)
	UA_Safari_MacOS = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Safari/605.1.15"
	// Safari usually doesn't send sec-ch-ua yet

	// Chrome (Android Mobile)
	UA_Chrome_Android       = "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36"
	SecChUa_Chrome_Mobile   = `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`
	SecChUaPlatform_Android = `"Android"`

	// Safari (iOS Mobile)
	UA_Safari_iOS = "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1"
)

// DefaultAcceptEncoding for browsers (Gzip, Deflate, Br)
const DefaultAcceptEncoding = "gzip, deflate, br"

// DefaultAcceptLanguage
const DefaultAcceptLanguage = "en-US,en;q=0.9"
