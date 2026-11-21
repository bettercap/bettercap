package zerogod

import (
	"crypto/x509"
	"testing"

	"github.com/bettercap/bettercap/v2/tls/emudhra"
)

func TestLoadTLSConfigEmudhra(t *testing.T) {
	sess := createMockSession()
	mod := NewZeroGod(sess)

	sess.Env.Set("zerogod.advertise.ca.mode", "emudhra")
	sess.Env.Set("zerogod.advertise.ca.emudhra", "auto")
	sess.Env.Set("zerogod.advertise.ca.chain", emudhra.ChainEmSign)

	cfg, err := mod.loadTLSConfig("iot.example.com")
	if err != nil {
		t.Fatalf("could not build the advertiser TLS config: %v", err)
	}

	if len(cfg.Certificates) != 1 {
		t.Fatalf("expected one certificate, got %d", len(cfg.Certificates))
	}

	cert := cfg.Certificates[0]
	if cert.Leaf == nil {
		t.Fatal("expected a parsed leaf")
	}

	roots := x509.NewCertPool()
	for _, anchor := range emudhra.Anchors(emudhra.ChainEmSign) {
		roots.AddCert(anchor)
	}

	intermediates := x509.NewCertPool()
	for _, raw := range cert.Certificate[1:] {
		parsed, err := x509.ParseCertificate(raw)
		if err != nil {
			t.Fatal(err)
		}
		intermediates.AddCert(parsed)
	}

	if _, err := cert.Leaf.Verify(x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		DNSName:       "iot.example.com",
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}); err != nil {
		t.Fatalf("advertised certificate does not verify: %v", err)
	}
}
