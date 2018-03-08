package modules

import (
	"net/http"
	"sync"
	// "strings"

	// "github.com/bettercap/bettercap/core"
	// "github.com/bettercap/bettercap/log"

	"github.com/elazarl/goproxy"
)

type cookieTracker struct {
	sync.RWMutex
	set map[string]string
}

func NewCookieTracker() *cookieTracker {
	return &cookieTracker{
		set: make(map[string]string),
	}
}

type SSLStripper struct {
	Enabled bool
	cookies *cookieTracker
}

func NewSSLStripper(enabled bool) *SSLStripper {
	return &SSLStripper{
		Enabled: enabled,
		cookies: NewCookieTracker(),
	}
}

// sslstrip preprocessing, takes care of:
//
// - patching / removing security related headers
// - making unknown session cookies expire
// - handling stripped domains
func (s *SSLStripper) Preprocess(req *http.Request, ctx *goproxy.ProxyCtx) (redir *http.Response) {
	if s.Enabled == false {
		return
	}

	return
}
