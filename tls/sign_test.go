package tls

import (
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"testing"
)

// newTestCA returns a CA tls.Certificate whose Certificate chain contains
// chainCerts certificates (the first one is the actual signing CA, the
// remaining ones simulate upper intermediates loaded from the proxy
// certificate file).
func newTestCA(t *testing.T, chainCerts int) tls.Certificate {
	t.Helper()

	priv, caDER, err := CreateCertificate(CertConfig{
		Bits:       2048,
		CommonName: "bettercap test CA",
	}, true)
	if err != nil {
		t.Fatalf("could not create test CA: %v", err)
	}

	chain := [][]byte{caDER}
	for i := 1; i < chainCerts; i++ {
		_, extraDER, err := CreateCertificate(CertConfig{
			Bits:       2048,
			CommonName: "bettercap test upper intermediate",
		}, true)
		if err != nil {
			t.Fatalf("could not create test intermediate: %v", err)
		}
		chain = append(chain, extraDER)
	}

	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatalf("could not parse test CA: %v", err)
	}

	return tls.Certificate{
		Certificate: chain,
		PrivateKey:  priv,
		Leaf:        caCert,
	}
}

// signForUnreachableHost signs a spoofed certificate for a host with no
// listener, so SignCertificateForHost takes its offline fallback template
// path instead of fetching a real server certificate. Port 1 on the
// loopback interface is refused immediately and keeps the test hermetic.
func signForUnreachableHost(t *testing.T, ca *tls.Certificate, bits int) *x509.Certificate {
	t.Helper()

	cert, err := SignCertificateForHost(ca, "127.0.0.1", 1, bits)
	if err != nil {
		t.Fatalf("could not sign spoofed certificate: %v", err)
	}

	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatalf("could not parse spoofed certificate: %v", err)
	}

	return leaf
}

func TestSignCertificateForHostKeySize(t *testing.T) {
	for _, bits := range []int{2048, 4096} {
		ca := newTestCA(t, 1)
		leaf := signForUnreachableHost(t, &ca, bits)

		pub, ok := leaf.PublicKey.(*rsa.PublicKey)
		if !ok {
			t.Fatalf("bits=%d: expected an RSA public key, got %T", bits, leaf.PublicKey)
		}

		if pub.N.BitLen() != bits {
			t.Errorf("expected a %d bit RSA key, got %d", bits, pub.N.BitLen())
		}

		caCert, err := x509.ParseCertificate(ca.Certificate[0])
		if err != nil {
			t.Fatalf("could not parse test CA: %v", err)
		}
		if err := leaf.CheckSignatureFrom(caCert); err != nil {
			t.Errorf("spoofed certificate is not signed by the CA: %v", err)
		}
	}
}

func TestSignCertificateForHostDefaultKeySize(t *testing.T) {
	for _, bits := range []int{0, -1} {
		ca := newTestCA(t, 1)
		leaf := signForUnreachableHost(t, &ca, bits)

		pub, ok := leaf.PublicKey.(*rsa.PublicKey)
		if !ok {
			t.Fatalf("bits=%d: expected an RSA public key, got %T", bits, leaf.PublicKey)
		}

		if pub.N.BitLen() != 2048 {
			t.Errorf("bits=%d: expected fallback to 2048 bits, got %d", bits, pub.N.BitLen())
		}
	}
}

func TestSignCertificateForHostServesFullChain(t *testing.T) {
	ca := newTestCA(t, 3)

	cert, err := SignCertificateForHost(&ca, "127.0.0.1", 1, 2048)
	if err != nil {
		t.Fatalf("could not sign spoofed certificate: %v", err)
	}

	if len(cert.Certificate) != len(ca.Certificate)+1 {
		t.Fatalf("expected %d certificates (leaf + full CA chain), got %d",
			len(ca.Certificate)+1, len(cert.Certificate))
	}

	for i, caDER := range ca.Certificate {
		if !bytes.Equal(cert.Certificate[i+1], caDER) {
			t.Errorf("chain position %d: served certificate does not match the CA file", i+1)
		}
	}
}
