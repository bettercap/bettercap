package tls

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/bettercap/bettercap/v2/session"
)

func TestCertConfigToModule(t *testing.T) {
	prefix := "test"
	defaults := DefaultLegitConfig

	dummyEnv, err := session.NewEnvironment("")
	if err != nil {
		t.Fatal(err)
	}
	dummySession := &session.Session{Env: dummyEnv}
	m := session.NewSessionModule(prefix, dummySession)

	CertConfigToModule(prefix, &m, defaults)

	// Check if parameters were added
	if len(m.Parameters()) != 6 {
		t.Errorf("expected 6 parameters, got %d", len(m.Parameters()))
	}
}

func TestCertConfigFromModule(t *testing.T) {
	dummyEnv, err := session.NewEnvironment("")
	if err != nil {
		t.Fatal(err)
	}
	dummySession := &session.Session{Env: dummyEnv}
	m := session.NewSessionModule("test", dummySession)
	prefix := "test"

	// Set some parameters
	m.AddParam(session.NewIntParameter(prefix+".certificate.bits", "2048", "dummy desc"))
	m.AddParam(session.NewStringParameter(prefix+".certificate.country", "TestCountry", ".*", "dummy desc"))
	m.AddParam(session.NewStringParameter(prefix+".certificate.locality", "TestLocality", ".*", "dummy desc"))
	m.AddParam(session.NewStringParameter(prefix+".certificate.organization", "TestOrg", ".*", "dummy desc"))
	m.AddParam(session.NewStringParameter(prefix+".certificate.organizationalunit", "TestUnit", ".*", "dummy desc"))
	m.AddParam(session.NewStringParameter(prefix+".certificate.commonname", "TestCN", ".*", "dummy desc"))

	cfg, err := CertConfigFromModule(prefix, m)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if cfg.Bits != 2048 || cfg.Country != "TestCountry" || cfg.Locality != "TestLocality" ||
		cfg.Organization != "TestOrg" || cfg.OrganizationalUnit != "TestUnit" || cfg.CommonName != "TestCN" {
		t.Error("config not parsed correctly")
	}
}

func TestCreateCertificate(t *testing.T) {
	cfg := DefaultLegitConfig
	cfg.Bits = 1024 // smaller for test

	priv, certBytes, err := CreateCertificate(cfg, true)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if priv == nil {
		t.Error("private key is nil")
	}
	if len(certBytes) == 0 {
		t.Error("cert bytes empty")
	}

	// Parse to verify
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		t.Errorf("could not parse cert: %v", err)
	}
	if cert.Subject.CommonName != cfg.CommonName {
		t.Errorf("common name mismatch: %s != %s", cert.Subject.CommonName, cfg.CommonName)
	}
	if !cert.IsCA {
		t.Error("not CA")
	}
}

func TestGenerate(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "tlstest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	certPath := filepath.Join(tempDir, "test.cert")
	keyPath := filepath.Join(tempDir, "test.key")

	cfg := DefaultLegitConfig
	cfg.Bits = 1024

	err = Generate(cfg, certPath, keyPath, false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check files exist
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		t.Error("cert file not created")
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Error("key file not created")
	}

	// Load and verify
	certBytes, _ := ioutil.ReadFile(certPath)
	keyBytes, _ := ioutil.ReadFile(keyPath)

	certBlock, _ := pem.Decode(certBytes)
	if certBlock == nil || certBlock.Type != "CERTIFICATE" {
		t.Error("invalid cert PEM")
	}

	keyBlock, _ := pem.Decode(keyBytes)
	if keyBlock == nil || keyBlock.Type != "RSA PRIVATE KEY" {
		t.Error("invalid key PEM")
	}

	priv, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		t.Errorf("invalid private key: %v", err)
	}
	if priv.N.BitLen() != 1024 {
		t.Errorf("key bits mismatch: %d", priv.N.BitLen())
	}
}
