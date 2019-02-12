package http_proxy

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/elazarl/goproxy"
	"github.com/jpillora/go-tld"
)

type CookieTracker struct {
	sync.RWMutex
	set map[string]bool
}

func NewCookieTracker() *CookieTracker {
	return &CookieTracker{
		set: make(map[string]bool),
	}
}

func (t *CookieTracker) domainOf(req *http.Request) string {
	if parsed, err := tld.Parse(req.Host); err != nil {
		return req.Host
	} else {
		return fmt.Sprintf("%s.%s", parsed.Domain, parsed.TLD)
	}
}

func (t *CookieTracker) keyOf(req *http.Request) string {
	client := strings.Split(req.RemoteAddr, ":")[0]
	domain := t.domainOf(req)
	return fmt.Sprintf("%s-%s", client, domain)
}

func (t *CookieTracker) IsClean(req *http.Request) bool {
	// we only clean GET requests
	if req.Method != "GET" {
		return true
	}

	// does the request have any cookie?
	cookie := req.Header.Get("Cookie")
	if cookie == "" {
		return true
	}

	t.RLock()
	defer t.RUnlock()

	// was it already processed?
	if _, found := t.set[t.keyOf(req)]; found {
		return true
	}

	// unknown session cookie
	return false
}

func (t *CookieTracker) Track(req *http.Request) {
	t.Lock()
	defer t.Unlock()
	t.set[t.keyOf(req)] = true
}

func (t *CookieTracker) Expire(req *http.Request) *http.Response {
	domain := t.domainOf(req)
	redir := goproxy.NewResponse(req, "text/plain", 302, "")

	for _, c := range req.Cookies() {
		redir.Header.Add("Set-Cookie", fmt.Sprintf("%s=EXPIRED; path=/; domain=%s; Expires=Mon, 01-Jan-1990 00:00:00 GMT", c.Name, domain))
		redir.Header.Add("Set-Cookie", fmt.Sprintf("%s=EXPIRED; path=/; domain=%s; Expires=Mon, 01-Jan-1990 00:00:00 GMT", c.Name, c.Domain))
	}

	redir.Header.Add("Location", fmt.Sprintf("http://%s/", req.Host))
	redir.Header.Add("Connection", "close")

	return redir
}
