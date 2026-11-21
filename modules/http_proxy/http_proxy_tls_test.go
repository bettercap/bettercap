package http_proxy

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bettercap/bettercap/v2/internal/certtest"
	btls "github.com/bettercap/bettercap/v2/tls"
	"github.com/bettercap/bettercap/v2/tls/emudhra"
)

// startMITMServer returns a TLS server that terminates connections with the
// certificate produced by TLSConfigFromCA and answers every request with
// "intercepted". It returns the listener address.
func startMITMServer(t *testing.T, proxy *HTTPProxy, ca *tls.Certificate, host string, port int) string {
	t.Helper()

	tlsConfig, err := proxy.TLSConfigFromCA(ca)(fmt.Sprintf("%s:%d", host, port), nil)
	if err != nil {
		t.Fatalf("could not build the MITM TLS config: %v", err)
	}

	listener, err := tls.Listen("tcp", "127.0.0.1:0", tlsConfig)
	if err != nil {
		t.Fatalf("could not listen: %v", err)
	}
	t.Cleanup(func() { listener.Close() })

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "intercepted")
		}),
	}
	go server.Serve(listener)
	t.Cleanup(func() { server.Close() })

	return listener.Addr().String()
}

// verifyMITMChain validates the chain served by the proxy without relying on
// x509.Verify, which distrusted the eMudhra SPKIs in Go 1.27.4.
func verifyMITMChain(rawCerts [][]byte, _ [][]*x509.Certificate) error {
	if len(rawCerts) == 0 {
		return fmt.Errorf("no certificates presented")
	}

	leaf, err := x509.ParseCertificate(rawCerts[0])
	if err != nil {
		return err
	}

	intermediates := make([]*x509.Certificate, 0, len(rawCerts)-1)
	for _, raw := range rawCerts[1:] {
		parsed, err := x509.ParseCertificate(raw)
		if err != nil {
			return err
		}
		intermediates = append(intermediates, parsed)
	}

	return certtest.VerifyChain(leaf, intermediates, emudhra.AllAnchors(), "127.0.0.1")
}

// TestEmudhraMITMInterceptsHTTPS verifies that the spoofed certificate chain
// produced for an intercepted host validates against the eMudhra root and that
// a client with an unrelated trust store rejects it.
func TestEmudhraMITMInterceptsHTTPS(t *testing.T) {
	// the origin certificate is cloned by SignCertificateForHost, so its
	// subject and SANs are reused for the spoofed leaf
	origin := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer origin.Close()
	originURL := origin.URL
	originPort := origin.Listener.Addr().(*net.TCPAddr).Port

	caCfg := btls.CAConfig{
		Mode:    "emudhra",
		Emudhra: "auto",
		Chain:   emudhra.ChainEmSign,
	}
	ca, err := caCfg.Load()
	if err != nil {
		t.Fatalf("could not load the eMudhra CA: %v", err)
	}

	sess, _ := createMockSession()
	proxy := NewHTTPProxy(sess, "test")
	proxy.LeafBits = 2048
	defer ClearCertCache()

	addr := startMITMServer(t, proxy, ca, "127.0.0.1", originPort)

	// Go 1.27.4+ hardcodes the compromised eMudhra SPKIs in its
	// distrusted list, so a normal Go client rejects the chain even when
	// the eMudhra root is trusted. Verify signatures and structure
	// manually to exercise the interception path.
	trusted := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify:    true,
				VerifyPeerCertificate: verifyMITMChain,
			},
		},
	}

	resp, err := trusted.Get("https://" + addr)
	if err != nil {
		t.Fatalf("interception failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "intercepted" {
		t.Fatalf("unexpected body %q (origin %s)", string(body), originURL)
	}

	// A client without the eMudhra root must reject the spoofed chain.
	untrusted := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: x509.NewCertPool()},
		},
	}
	if _, err := untrusted.Get("https://" + addr); err == nil {
		t.Fatal("expected a client without the eMudhra root to reject the forged certificate")
	}
}

// TestTLSConfigFromCAReportsClientCapabilities checks that the MITM config
// exposes the callbacks used to log client capabilities and the handshake
// verdict.
func TestTLSConfigFromCAReportsClientCapabilities(t *testing.T) {
	caCfg := btls.CAConfig{Mode: "emudhra", Emudhra: "auto", Chain: emudhra.ChainAAA}
	ca, err := caCfg.Load()
	if err != nil {
		t.Fatal(err)
	}

	sess, _ := createMockSession()
	proxy := NewHTTPProxy(sess, "test")
	proxy.LeafBits = 2048

	config, err := proxy.TLSConfigFromCA(ca)("127.0.0.1:1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if config.GetConfigForClient == nil {
		t.Fatal("expected a GetConfigForClient callback")
	}
	if config.VerifyConnection == nil {
		t.Fatal("expected a VerifyConnection callback")
	}
}

// TestConfigureTLSClearsCertCache checks that reconfiguring the proxy with a
// different CA invalidates the cached spoofed certificates.
func TestConfigureTLSClearsCertCache(t *testing.T) {
	sess, _ := createMockSession()
	proxy := NewHTTPProxy(sess, "test")

	caCfg := btls.CAConfig{Mode: "emudhra", Emudhra: "auto", Chain: emudhra.ChainEmSign}
	ca, err := caCfg.Load()
	if err != nil {
		t.Fatal(err)
	}

	setCachedCert("iot.example.com", 443, ca)
	if getCachedCert("iot.example.com", 443) == nil {
		t.Fatal("expected the certificate to be cached")
	}

	if err := proxy.ConfigureTLS("127.0.0.1", 0, 0, false, "", ca, "", false, 2048); err != nil {
		t.Fatal(err)
	}

	if getCachedCert("iot.example.com", 443) != nil {
		t.Fatal("expected the certificate cache to be cleared")
	}
	if proxy.CA != ca {
		t.Fatal("expected the proxy to store the configured CA")
	}
}

func TestEmudhraRevocationHandling(t *testing.T) {
	caCfg := btls.CAConfig{Mode: "emudhra", Emudhra: "auto", Chain: emudhra.ChainEmSign}
	ca, err := caCfg.Resolve()
	if err != nil {
		t.Fatal(err)
	}

	cert, err := caCfg.LoadForHost("iot.example.com", 2048)
	if err != nil {
		t.Fatal(err)
	}

	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(leaf.CRLDistributionPoints) != 1 {
		t.Fatalf("expected one CRL distribution point, got %v", leaf.CRLDistributionPoints)
	}

	sess, _ := createMockSession()
	proxy := NewHTTPProxy(sess, "test")

	crlReq := httptest.NewRequest(http.MethodGet, leaf.CRLDistributionPoints[0], nil)
	_, resp := proxy.onRequestFilter(crlReq, nil)
	if resp == nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected the synthetic CRL, got %+v", resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	crl, err := x509.ParseRevocationList(body)
	if err != nil {
		t.Fatalf("could not parse the served CRL: %v", err)
	}
	if err := crl.CheckSignatureFrom(ca.Cert); err != nil {
		t.Fatalf("CRL signature does not verify: %v", err)
	}

	ocspReq := httptest.NewRequest(http.MethodPost, "http://ocsp-a.emsign.com/", nil)
	_, resp = proxy.onRequestFilter(ocspReq, nil)
	if resp == nil || resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("expected revocation lookups to be blocked, got %+v", resp)
	}

	normalReq := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	if _, resp = proxy.onRequestFilter(normalReq, nil); resp != nil {
		t.Fatalf("expected normal requests to be proxied, got %+v", resp)
	}
}
