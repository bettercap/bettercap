package emudhra

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"
)

var expectedFingerprints = map[string]string{
	"emSign/EM DV TLS CA - G3B": "42C4ACCD12C4F59597925E1A08A7A90102608543809308D50EC427E118BA1493",
	"emSign/EM DV TLS CA - G3C": "0164D7CEAE5BB909C272AF151FF44B1177AE9B47D9424B94F1D9558C30934D19",
	"emSign/EM EV TLS CA - G3B": "E7568FE6E12331F4FD7DE61DA4C7559D0F8EB192A218D4741A3BA75184333838",
	"emSign/EM EV TLS CA - G3C": "432EC25F0D28DE9006D97BF752F17BBB2CD41F669933FCB072D9A1148D0261F3",
	"emSign/EM OV TLS CA - G3B": "98B896625C0C230ADEEC3EB2CF6AAD5001ECF9156A68644E82374CE5833BC25B",
	"emSign/EM OV TLS CA - G3C": "CB6E15165AC802E9B3F993F16C25D41F4E81C21EBDE71AE1B012AC00BF9C7C3F",
	"aaa/EM DV TLS CA - G3B":    "45A883F349AEDFD5F3A649B5498FFB121712BA8B448F3F2A91DBC5A766FCDDBE",
	"aaa/EM DV TLS CA - G3C":    "62B7F9868263F72F6911FC78CEAF72C1979158B7368BF648683BB2AA39B02E72",
	"aaa/EM EV TLS CA - G3B":    "579480AEA166290BD3C643589E795FC0D0E54E86505D180F1DFF99EFD3199512",
	"aaa/EM EV TLS CA - G3C":    "FD9D5C2DBBD4995AB69B8D50C439D9527B00529199C2FD01B2174EB58F07021F",
	"aaa/EM OV TLS CA - G3B":    "5F1DED4DD590AF74631290CF073B122A2AAC6A1FA479778A0BD43579BD832699",
	"aaa/EM OV TLS CA - G3C":    "7485B51C0ED192878FA1D2F948FB529AE1842CC5086C4B67F7BE5B2039DE9489",
}

var expectedKeys = map[string]int{
	"EM DV TLS CA - G3B": 156,
	"EM DV TLS CA - G3C": 230,
	"EM EV TLS CA - G3B": 115,
	"EM EV TLS CA - G3C": 56,
	"EM OV TLS CA - G3B": 155,
	"EM OV TLS CA - G3C": 4,
}

func TestEmbeddedFingerprints(t *testing.T) {
	if err := Err(); err != nil {
		t.Fatalf("could not load embedded certificates: %v", err)
	}

	seen := map[string]bool{}
	for _, ca := range List() {
		key := ca.Chain + "/" + ca.Name
		want, found := expectedFingerprints[key]
		if !found {
			t.Errorf("unexpected embedded CA %q", key)
			continue
		}
		if ca.SHA256 != want {
			t.Errorf("%s: fingerprint is %s, expected %s", key, ca.SHA256, want)
		}
		seen[key] = true
	}

	for key := range expectedFingerprints {
		if !seen[key] {
			t.Errorf("missing embedded CA %q", key)
		}
	}
}

func TestRecoveredKeys(t *testing.T) {
	for _, ca := range List() {
		want, found := expectedKeys[ca.Name]
		if !found {
			t.Fatalf("no expected key for %s", ca.Name)
		}
		got := ca.KeyValue()
		if got != want {
			t.Errorf("%s: recovered key is %d, expected %d", ca.Name, got, want)
		}
	}
}

func TestChainsVerify(t *testing.T) {
	anchors := AllAnchors()
	if len(anchors) != 3 {
		t.Fatalf("expected 3 trust anchors, got %d", len(anchors))
	}

	for _, chain := range Chains() {
		ca, err := Find("auto", chain)
		if err != nil {
			t.Fatalf("%s: %v", chain, err)
		}

		tlsCert, err := ca.TLSCertificate(false)
		if err != nil {
			t.Fatalf("%s: %v", chain, err)
		}

		// A client receives the leaf plus both chain variants: the
		// compromised intermediate, the cross-signed emSign Root TLS
		// CA - G3 and the AAA cross-sign.
		if len(tlsCert.Certificate) != 3 {
			t.Fatalf("%s: expected 3 chain certificates, got %d", chain, len(tlsCert.Certificate))
		}

		intermediates := x509.NewCertPool()
		for _, raw := range tlsCert.Certificate {
			cert, err := x509.ParseCertificate(raw)
			if err != nil {
				t.Fatalf("%s: %v", chain, err)
			}
			intermediates.AddCert(cert)
		}

		leaf := signTestLeaf(t, ca)
		for _, anchor := range anchors {
			roots := x509.NewCertPool()
			roots.AddCert(anchor)

			if _, err := leaf.Verify(x509.VerifyOptions{
				Roots:         roots,
				Intermediates: intermediates,
				DNSName:       "iot.example.com",
				KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			}); err != nil {
				t.Errorf("%s: verification against anchor %q failed: %v", chain, anchor.Subject.CommonName, err)
			}
		}
	}
}

func TestServedChainOrder(t *testing.T) {
	// OpenSSL only follows the first intermediate with a matching subject,
	// so the preferred variant must come first, followed by the other one.
	for _, tc := range []struct {
		chain   string
		issuers []string
	}{
		{
			chain: ChainEmSign,
			issuers: []string{
				"emSign Root TLS CA - G3",  // emSign-issued intermediate
				"emSign ECC Root CA - G3",  // cross-signed Root TLS CA
				"AAA Certificate Services", // AAA cross-sign
			},
		},
		{
			chain: ChainAAA,
			issuers: []string{
				"AAA Certificate Services", // AAA cross-sign first
				"emSign Root TLS CA - G3",  // emSign-issued intermediate
				"emSign ECC Root CA - G3",  // cross-signed Root TLS CA
			},
		},
	} {
		t.Run(tc.chain, func(t *testing.T) {
			ca, err := Find("auto", tc.chain)
			if err != nil {
				t.Fatal(err)
			}

			cert, err := ca.TLSCertificate(false)
			if err != nil {
				t.Fatal(err)
			}

			if len(cert.Certificate) != len(tc.issuers) {
				t.Fatalf("expected %d certificates, got %d", len(tc.issuers), len(cert.Certificate))
			}

			for i, raw := range cert.Certificate {
				parsed, err := x509.ParseCertificate(raw)
				if err != nil {
					t.Fatal(err)
				}
				if parsed.Issuer.CommonName != tc.issuers[i] {
					t.Fatalf("certificate %d: expected issuer %q, got %q",
						i, tc.issuers[i], parsed.Issuer.CommonName)
				}
			}
		})
	}
}

func TestChainVariantsShareKey(t *testing.T) {
	for _, ca := range List() {
		other := ca.counterpart()
		if other == nil {
			t.Fatalf("%s: missing %s counterpart", ca.Name, ca.Chain)
		}
		if ca.KeyValue() != other.KeyValue() {
			t.Fatalf("%s: key mismatch between chains (%d != %d)", ca.Name, ca.KeyValue(), other.KeyValue())
		}
	}
}

func TestFind(t *testing.T) {
	ca, err := Find("EM OV TLS CA - G3C", ChainEmSign)
	if err != nil {
		t.Fatal(err)
	}
	if ca.Name != "EM OV TLS CA - G3C" || ca.KeyValue() != 4 {
		t.Fatalf("unexpected CA: %s", ca)
	}

	byFingerprint, err := Find("42c4accd12c4", ChainEmSign)
	if err != nil {
		t.Fatal(err)
	}
	if byFingerprint.Name != "EM DV TLS CA - G3B" {
		t.Fatalf("unexpected CA: %s", byFingerprint)
	}

	if _, err := Find("EM OV TLS CA - G3C", ChainAAA); err != nil {
		t.Fatalf("expected the cross-signed variant to exist: %v", err)
	}

	if _, err := Find("does not exist", ChainEmSign); err == nil {
		t.Fatal("expected an error for an unknown CA")
	}
}

func signTestLeaf(t *testing.T, ca *CA) *x509.Certificate {
	t.Helper()

	key, err := ca.Recover()
	if err != nil {
		t.Fatal(err)
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "iot.example.com"},
		DNSNames:     []string{"iot.example.com"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	der, err := x509.CreateCertificate(rand.Reader, template, ca.Cert, &priv.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}

	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return leaf
}

func TestCRL(t *testing.T) {
	ca, err := Find("auto", ChainEmSign)
	if err != nil {
		t.Fatal(err)
	}

	der, err := ca.CRL()
	if err != nil {
		t.Fatalf("could not create the CRL: %v", err)
	}

	crl, err := x509.ParseRevocationList(der)
	if err != nil {
		t.Fatalf("could not parse the CRL: %v", err)
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
	if ca.CRLPath() != "/"+ca.CRLToken() {
		t.Fatalf("unexpected CRL path %q", ca.CRLPath())
	}
}

func TestCRLDistributionPoint(t *testing.T) {
	ca, err := Find("auto", ChainAAA)
	if err != nil {
		t.Fatal(err)
	}

	origin := &x509.Certificate{CRLDistributionPoints: []string{"http://crl.example.com/foo.crl"}}
	if got, want := ca.CRLDistributionPoint(origin), "http://crl.example.com"+ca.CRLPath(); got != want {
		t.Fatalf("got %q, expected %q", got, want)
	}

	if got, want := ca.CRLDistributionPoint(nil), "http://crl.emsign.com"+ca.CRLPath(); got != want {
		t.Fatalf("got %q, expected %q", got, want)
	}
}

func BenchmarkRecover(b *testing.B) {
	ca, err := Find("EM OV TLS CA - G3C", ChainEmSign)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		if _, err := recoverKey(ca.Cert); err != nil {
			b.Fatal(err)
		}
	}
}
