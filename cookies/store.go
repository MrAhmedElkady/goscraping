package cookies

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"

	"golang.org/x/net/publicsuffix"
)

// SessionJar implements http.CookieJar with session support
type SessionJar struct {
	jar http.CookieJar
	mu  sync.RWMutex
}

// NewSessionJar creates a new session jar
func NewSessionJar() *SessionJar {
	// Use standard net/http/cookiejar with publicsuffix
	jar, _ := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	return &SessionJar{
		jar: jar,
	}
}

// SetCookies handles the receipt of the cookies in a reply for the given URL
func (j *SessionJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.jar.SetCookies(u, cookies)
}

// Cookies returns the cookies to send in a request for the given URL
func (j *SessionJar) Cookies(u *url.URL) []*http.Cookie {
	return j.jar.Cookies(u)
}
