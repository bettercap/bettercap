package zerogod

import (
	"crypto/x509"
	"testing"

	"github.com/bettercap/bettercap/v2/internal/certtest"
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

	intermediates := []*x509.Certificate{}
	for _, raw := range cert.Certificate[1:] {
		parsed, err := x509.ParseCertificate(raw)
		if err != nil {
			t.Fatal(err)
		}
		intermediates = append(intermediates, parsed)
	}

	// Go 1.27.4+ distrusted eMudhra SPKIs by hash, so x509.Verify cannot
	// be used even though the chain is cryptographically valid.
	if err := certtest.VerifyChain(cert.Leaf, intermediates, emudhra.AllAnchors(), "iot.example.com"); err != nil {
		t.Fatalf("advertised certificate does not verify: %v", err)
	}
}
