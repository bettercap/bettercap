// Package certtest provides certificate chain verification helpers for tests.
//
// Go 1.27.4 added a hardcoded blocklist of distrusted SPKI hashes that
// includes the compromised eMudhra intermediates. x509.Verify rejects those
// chains regardless of the configured roots, so tests that only need to prove
// that a chain is structurally and cryptographically valid use this package
// instead.
package certtest

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"time"
)

// VerifyChain checks that leaf chains to one of roots through the given
// intermediates, that every certificate is currently valid, and that the leaf
// matches dnsName and is valid for server authentication.
func VerifyChain(leaf *x509.Certificate, intermediates, roots []*x509.Certificate, dnsName string) error {
	if leaf == nil {
		return fmt.Errorf("nil leaf certificate")
	}

	if err := leaf.VerifyHostname(dnsName); err != nil {
		return err
	}
	if !hasServerAuth(leaf) {
		return fmt.Errorf("%s is not valid for server authentication", leaf.Subject.CommonName)
	}

	now := time.Now()
	if err := checkValidity(leaf, now); err != nil {
		return err
	}

	var walk func(cert *x509.Certificate, depth int) error
	walk = func(cert *x509.Certificate, depth int) error {
		if depth > 10 {
			return fmt.Errorf("chain is too long")
		}

		for _, root := range roots {
			if bytes.Equal(cert.Raw, root.Raw) {
				return nil
			}
			if !bytes.Equal(cert.RawIssuer, root.RawSubject) {
				continue
			}
			if err := cert.CheckSignatureFrom(root); err != nil {
				continue
			}
			return checkValidity(root, now)
		}

		for _, parent := range intermediates {
			if bytes.Equal(parent.Raw, cert.Raw) {
				continue
			}
			if !bytes.Equal(cert.RawIssuer, parent.RawSubject) {
				continue
			}
			if !parent.IsCA {
				continue
			}
			if err := cert.CheckSignatureFrom(parent); err != nil {
				continue
			}
			if err := checkValidity(parent, now); err != nil {
				continue
			}
			if err := walk(parent, depth+1); err == nil {
				return nil
			}
		}

		return fmt.Errorf("no path from %q to a trusted root", cert.Subject.CommonName)
	}

	return walk(leaf, 0)
}

func checkValidity(cert *x509.Certificate, now time.Time) error {
	if now.Before(cert.NotBefore) {
		return fmt.Errorf("%s is not valid yet", cert.Subject.CommonName)
	}
	if now.After(cert.NotAfter) {
		return fmt.Errorf("%s has expired", cert.Subject.CommonName)
	}
	return nil
}

func hasServerAuth(cert *x509.Certificate) bool {
	if len(cert.ExtKeyUsage) == 0 {
		return true
	}
	for _, eku := range cert.ExtKeyUsage {
		if eku == x509.ExtKeyUsageAny || eku == x509.ExtKeyUsageServerAuth {
			return true
		}
	}
	return false
}
