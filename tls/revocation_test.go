package tls

import (
	"crypto/x509"
	"io"
	"net/http"
	"testing"

	"github.com/bettercap/bettercap/v2/tls/emudhra"
)

func TestRevocationHandlerServesCRL(t *testing.T) {
	cfg := CAConfig{Mode: "emudhra", Emudhra: "auto", Chain: emudhra.ChainEmSign}
	ca, err := cfg.Resolve()
	if err != nil {
		t.Fatal(err)
	}

	cert, err := cfg.LoadForHost("iot.example.com", 2048)
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

	resp, handled := HandleRevocationRequest(mustRequest(t, leaf.CRLDistributionPoints[0]))
	if !handled {
		t.Fatal("expected the CRL request to be handled")
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d", resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != "application/pkix-crl" {
		t.Fatalf("unexpected content type %q", resp.Header.Get("Content-Type"))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	crl, err := x509.ParseRevocationList(body)
	if err != nil {
		t.Fatalf("could not parse the served CRL: %v", err)
	}
	if crl.Issuer.String() != ca.Cert.Subject.String() {
		t.Fatalf("CRL issuer is %q, expected %q", crl.Issuer, ca.Cert.Subject)
	}
	if err := crl.CheckSignatureFrom(ca.Cert); err != nil {
		t.Fatalf("CRL signature does not verify: %v", err)
	}
	if len(crl.RevokedCertificateEntries) != 0 {
		t.Fatalf("expected an empty CRL, got %d entries", len(crl.RevokedCertificateEntries))
	}
	if !crl.NextUpdate.After(crl.ThisUpdate) {
		t.Fatal("expected NextUpdate after ThisUpdate")
	}
}

func TestRevocationHandlerBlocksKnownHosts(t *testing.T) {
	cfg := CAConfig{Mode: "emudhra", Emudhra: "auto", Chain: emudhra.ChainAAA}
	if _, err := cfg.Resolve(); err != nil {
		t.Fatal(err)
	}

	for host := range revocationHosts {
		for _, scheme := range []string{"http", "https"} {
			req := mustRequest(t, scheme+"://"+host+"/some/path")
			resp, handled := HandleRevocationRequest(req)
			if !handled {
				t.Errorf("%s: expected the request to be handled", host)
				continue
			}
			if resp.StatusCode != http.StatusBadGateway {
				t.Errorf("%s: expected status %d, got %d", host, http.StatusBadGateway, resp.StatusCode)
			}
		}
	}
}

func TestRevocationHandlerIgnoresUnrelatedRequests(t *testing.T) {
	cfg := CAConfig{Mode: "emudhra", Emudhra: "auto", Chain: emudhra.ChainEmSign}
	if _, err := cfg.Resolve(); err != nil {
		t.Fatal(err)
	}

	for _, url := range []string{
		"http://example.com/",
		"https://example.com/ocsp",
		"http://crl.example.com/some.crl",
	} {
		if _, handled := HandleRevocationRequest(mustRequest(t, url)); handled {
			t.Errorf("%s: expected the request to be proxied normally", url)
		}
	}
}

func TestSignCertificateForHostAddsEmudhraCRL(t *testing.T) {
	cfg := CAConfig{Mode: "emudhra", Emudhra: "auto", Chain: emudhra.ChainEmSign}
	ca, err := cfg.Load()
	if err != nil {
		t.Fatal(err)
	}

	leaf := signForUnreachableHost(t, ca, 2048)
	if len(leaf.CRLDistributionPoints) != 1 {
		t.Fatalf("expected one CRL distribution point, got %v", leaf.CRLDistributionPoints)
	}

	req := mustRequest(t, leaf.CRLDistributionPoints[0])
	resp, handled := HandleRevocationRequest(req)
	if !handled || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected the synthetic CRL, got handled=%v status=%v", handled, resp)
	}
}

func mustRequest(t *testing.T, url string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("could not build request: %v", err)
	}
	return req
}
