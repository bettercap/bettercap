package syn_scan

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"strings"
)

func isTitleElement(n *html.Node) bool {
	return n.Type == html.ElementNode && strings.ToLower(n.Data) == "title"
}

func searchForTitle(n *html.Node) string {
	if isTitleElement(n) && n.FirstChild != nil {
		return n.FirstChild.Data
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := searchForTitle(c); result != "" {
			return result
		}
	}

	return ""
}

func httpGrabber(mod *SynScanner, ip string, port int) string {
	schema := "http"
	client := &http.Client{
		Timeout: bannerGrabTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	sport := fmt.Sprintf("%d", port)
	if strings.Contains(sport, "443") {
		schema = "https"
		client = &http.Client{
			Timeout: bannerGrabTimeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
					VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
						return nil
					},
				},
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return nil
			},
		}
	}

	// https://stackoverflow.com/questions/12260003/connect-returns-invalid-argument-with-ipv6-address
	if strings.Contains(ip, ":") {
		ip = fmt.Sprintf("[%s%%25%s]", ip, mod.Session.Interface.Name())
	}

	url := fmt.Sprintf("%s://%s:%d/", schema, ip, port)
	resp, err := client.Get(url)
	if err != nil {
		mod.Debug("error while grabbing banner from %s: %v", url, err)
		return ""
	}
	defer resp.Body.Close()

	fallback := ""
	for name, values := range resp.Header {
		for _, value := range values {
			header := strings.ToLower(name)
			if len(value) > len(fallback) && (header == "x-powered-by" || header == "server") {
				mod.Debug("found header %s for %s:%d -> %s", header, ip, port, value)
				fallback = value
			}
		}
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		mod.Debug("error while reading and parsing response from %s: %v", url, err)
		return fallback
	}

	if title := searchForTitle(doc); title != "" {
		return title
	}

	return fallback
}
