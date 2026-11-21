package tls

import (
	stdtls "crypto/tls"
	"crypto/x509"
	"path/filepath"
	"testing"

	"github.com/bettercap/bettercap/v2/internal/certtest"
	"github.com/bettercap/bettercap/v2/tls/emudhra"
)

func verifyServedChain(t *testing.T, cert *stdtls.Certificate, chain string, includeRoot bool) {
	t.Helper()

	intermediates := []*x509.Certificate{}
	for _, raw := range cert.Certificate[1:] {
		parsed, err := x509.ParseCertificate(raw)
		if err != nil {
			t.Fatalf("could not parse chain certificate: %v", err)
		}
		intermediates = append(intermediates, parsed)
	}

	// both chain variants are always sent: intermediate, cross-signed
	// Root TLS CA and the AAA cross-sign
	expected := 4
	if includeRoot {
		expected = 5
	}
	if len(cert.Certificate) != expected {
		t.Fatalf("%s: expected %d certificates, got %d", chain, expected, len(cert.Certificate))
	}

	leaf := cert.Leaf
	if leaf == nil {
		var err error
		if leaf, err = x509.ParseCertificate(cert.Certificate[0]); err != nil {
			t.Fatalf("could not parse leaf: %v", err)
		}
	}

	// Go 1.27.4+ distrusted eMudhra SPKIs by hash, so x509.Verify cannot
	// be used here even though the chain is cryptographically valid.
	if err := certtest.VerifyChain(leaf, intermediates, emudhra.AllAnchors(), "iot.example.com"); err != nil {
		t.Fatalf("served chain does not verify for %s: %v", chain, err)
	}

	if includeRoot {
		anchor := emudhra.Anchors(chain)[0]
		last, err := x509.ParseCertificate(cert.Certificate[len(cert.Certificate)-1])
		if err != nil {
			t.Fatal(err)
		}
		if last.Subject.CommonName != anchor.Subject.CommonName {
			t.Fatalf("expected root %q at the end of the chain, got %q", anchor.Subject.CommonName, last.Subject.CommonName)
		}
	}
}

func TestCAConfigLoadForHostEmudhraChains(t *testing.T) {
	for _, chain := range emudhra.Chains() {
		t.Run(chain, func(t *testing.T) {
			cfg := CAConfig{
				Mode:    "emudhra",
				Emudhra: "auto",
				Chain:   chain,
			}

			cert, err := cfg.LoadForHost("iot.example.com", 2048)
			if err != nil {
				t.Fatalf("could not sign leaf: %v", err)
			}

			if _, ok := cert.PrivateKey.(interface{}); !ok {
				t.Fatal("expected a private key")
			}

			verifyServedChain(t, cert, chain, false)
			if !cfg.IsEmudhra() {
				t.Fatal("expected IsEmudhra to be true")
			}
			if cfg.Describe() == "" {
				t.Fatal("expected a non empty description")
			}
		})
	}
}

func TestCAConfigIncludeRoot(t *testing.T) {
	cfg := CAConfig{
		Mode:        "emudhra",
		Emudhra:     "auto",
		Chain:       emudhra.ChainEmSign,
		IncludeRoot: true,
	}

	cert, err := cfg.LoadForHost("iot.example.com", 2048)
	if err != nil {
		t.Fatal(err)
	}
	verifyServedChain(t, cert, emudhra.ChainEmSign, true)
}

func TestCAConfigLoadEmudhraCA(t *testing.T) {
	cfg := CAConfig{Mode: "emudhra", Emudhra: "EM OV TLS CA - G3C", Chain: emudhra.ChainEmSign}

	cert, err := cfg.Load()
	if err != nil {
		t.Fatal(err)
	}

	if cert.Leaf == nil || cert.Leaf.Subject.CommonName != "EM OV TLS CA - G3C" {
		t.Fatalf("unexpected CA certificate: %+v", cert.Leaf)
	}
	if _, ok := cert.PrivateKey.(interface{}); !ok {
		t.Fatal("expected a recovered private key")
	}
}

func TestCAConfigAutoGeneratesAndReloads(t *testing.T) {
	dir := t.TempDir()
	cfg := CAConfig{
		Mode:     "auto",
		CertFile: filepath.Join(dir, "ca.cert.pem"),
		KeyFile:  filepath.Join(dir, "ca.key.pem"),
		Config: CertConfig{
			Bits:       2048,
			CommonName: "bettercap test CA",
		},
		IsCA: true,
	}

	if cfg.FilesExist() {
		t.Fatal("expected no certificate files yet")
	}

	generated, err := cfg.Load()
	if err != nil {
		t.Fatalf("could not generate CA: %v", err)
	}
	if !cfg.FilesExist() {
		t.Fatal("expected the certificate files to be written")
	}

	reloaded, err := cfg.Load()
	if err != nil {
		t.Fatalf("could not reload CA: %v", err)
	}

	if generated.Leaf == nil || reloaded.Leaf == nil {
		t.Fatal("expected parsed leaf certificates")
	}
	if generated.Leaf.Subject.CommonName != reloaded.Leaf.Subject.CommonName {
		t.Fatal("reloaded a different certificate")
	}

	loaded, err := LoadCertificate(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Leaf == nil {
		t.Fatal("expected LoadCertificate to populate the leaf")
	}
}

func TestCAConfigErrors(t *testing.T) {
	if _, err := (CAConfig{Mode: "emudhra", Chain: "nope"}).Resolve(); err == nil {
		t.Fatal("expected an error for an unknown chain")
	}
	if _, err := (CAConfig{Mode: "wat"}).Load(); err == nil {
		t.Fatal("expected an error for an unknown mode")
	}
	if _, err := (CAConfig{Mode: "auto"}).Load(); err == nil {
		t.Fatal("expected an error when no files are configured")
	}
	if (CAConfig{Mode: "auto"}).IsEmudhra() {
		t.Fatal("auto is not emudhra")
	}
	if _, err := (CAConfig{Mode: "auto"}).Resolve(); err == nil {
		t.Fatal("expected Resolve to fail for auto mode")
	}
}
