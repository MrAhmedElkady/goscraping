package types

// Browser enum
type Browser string

const (
	BrowserAny     Browser = "any"
	BrowserChrome  Browser = "chrome"
	BrowserSafari  Browser = "safari"
	BrowserFirefox Browser = "firefox"
)

// Device enum
type Device string

const (
	DeviceAny     Device = "any"
	DeviceDesktop Device = "desktop"
	DeviceMobile  Device = "mobile"
	DeviceTablet  Device = "tablet"
)

// OS enum
type OS string

const (
	OSAny     OS = "any"
	OSWindows OS = "windows"
	OSMacOS   OS = "macos"
	OSLinux   OS = "linux"
	OSAndroid OS = "android"
	OSiOS     OS = "ios"
)
