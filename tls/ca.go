package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	"github.com/bettercap/bettercap/v2/session"
	"github.com/bettercap/bettercap/v2/tls/emudhra"
	"github.com/evilsocket/islazy/fs"
)

// CAConfig describes how a module should obtain the certificate authority it
// uses to sign or serve TLS certificates.
type CAConfig struct {
	// Mode is "auto" (generate or load CertFile/KeyFile) or "emudhra" (use one
	// of the compromised eMudhra intermediate CAs).
	Mode string
	// Emudhra selects the CA: "auto", a common name or a fingerprint prefix.
	Emudhra string
	// Chain is either emudhra.ChainEmSign or emudhra.ChainAAA.
	Chain string
	// IncludeRoot adds the trust anchor to the served chain.
	IncludeRoot bool
	// Host overrides the hostname used when signing a server certificate.
	Host string
	// CertFile and KeyFile are used in "auto" mode.
	CertFile string
	KeyFile  string
	// Config is used to generate certificates in "auto" mode.
	Config CertConfig
	// IsCA marks the generated certificate as a CA.
	IsCA bool
}

// CAConfigToModule registers the CA selection parameters for a module.
func CAConfigToModule(prefix string, m *session.SessionModule) {
	m.AddParam(session.NewStringParameter(prefix+".ca.mode", "auto", "(auto|emudhra)",
		"CA source: 'auto' (default) uses/generates the configured certificate and key, which you must install on the client; "+
			"'emudhra' signs with a revoked eMudhra intermediate, so it only works on clients with a stale trust store and no hard-fail CRL/OCSP or CT checks."))

	m.AddParam(session.NewStringParameter(prefix+".ca.emudhra", "auto", ".*",
		"Compromised CA used in 'emudhra' mode: 'auto' picks one, or give a common name or SHA-256 prefix. "+
			"The choice rarely matters; use a specific one to reproduce a known key leak or to test fingerprint pinning."))

	m.AddParam(session.NewStringParameter(prefix+".ca.chain", emudhra.ChainEmSign, "(emSign|aaa)",
		"Chain preference in 'emudhra' mode: 'emSign' (default) targets modern trust stores that have emSign ECC Root CA - G3; "+
			"'aaa' targets old or frozen stores that only trust AAA Certificate Services. Both chains are sent, but OpenSSL follows only the first, so pick 'aaa' for old OpenSSL clients."))

	m.AddParam(session.NewBoolParameter(prefix+".ca.include_root", "false",
		"Append the trust anchor to the served chain. Leave false; enable only if a client cannot build the chain without receiving the root."))

	m.AddParam(session.NewStringParameter(prefix+".ca.host", "", ".*",
		"Hostname for the fixed server certificate of https.server, dns.proxy and zerogod.advertise. "+
			"Set it to the name clients actually connect to (e.g. 'host.local' or 'cloudflare-dns.com'); empty derives it from the bind address or certificate common name."))
}

// CAConfigFromModule reads the CA selection parameters of a module.
func CAConfigFromModule(prefix string, m session.SessionModule) (cfg CAConfig, err error) {
	if err, cfg.Mode = m.StringParam(prefix + ".ca.mode"); err != nil {
		return
	}
	if err, cfg.Emudhra = m.StringParam(prefix + ".ca.emudhra"); err != nil {
		return
	}
	if err, cfg.Chain = m.StringParam(prefix + ".ca.chain"); err != nil {
		return
	}
	if err, cfg.IncludeRoot = m.BoolParam(prefix + ".ca.include_root"); err != nil {
		return
	}
	if err, cfg.Host = m.StringParam(prefix + ".ca.host"); err != nil {
		return
	}
	if err, cfg.CertFile = m.StringParam(prefix + ".certificate"); err != nil {
		return
	}
	if cfg.CertFile, err = fs.Expand(cfg.CertFile); err != nil {
		return
	}
	if err, cfg.KeyFile = m.StringParam(prefix + ".key"); err != nil {
		return
	}
	if cfg.KeyFile, err = fs.Expand(cfg.KeyFile); err != nil {
		return
	}
	cfg.Config, err = CertConfigFromModule(prefix, m)
	return
}

// IsEmudhra returns true when the config selects a compromised eMudhra CA.
func (c CAConfig) IsEmudhra() bool {
	return strings.EqualFold(strings.TrimSpace(c.Mode), "emudhra")
}

// FilesExist reports whether the local pair can be loaded. Generation is safe
// only when both paths are absent.
func (c CAConfig) FilesExist() (bool, error) {
	if c.CertFile == "" || c.KeyFile == "" {
		return false, fmt.Errorf("no certificate or key file configured")
	}
	certExists, err := regularFileExists(c.CertFile)
	if err != nil {
		return false, err
	}
	keyExists, err := regularFileExists(c.KeyFile)
	if err != nil {
		return false, err
	}
	if certExists != keyExists {
		return false, fmt.Errorf("incomplete TLS pair: certificate %q and key %q must both exist", c.CertFile, c.KeyFile)
	}
	return certExists, nil
}

// Resolve looks up the selected CA by its public certificate.
func (c CAConfig) Resolve() (*emudhra.CA, error) {
	if !c.IsEmudhra() {
		return nil, fmt.Errorf("CA mode is %q, not emudhra", c.Mode)
	}

	return emudhra.Find(c.Emudhra, c.Chain)
}

// Describe returns a human readable description of the selected CA.
func (c CAConfig) Describe() string {
	if c.IsEmudhra() {
		ca, err := c.Resolve()
		if err != nil {
			return "eMudhra: " + err.Error()
		}
		return "eMudhra CA " + ca.String()
	}
	return "local CA " + c.CertFile
}

// Load returns the CA certificate described by the config. In "auto" mode it
// loads CertFile/KeyFile, generating them if they do not exist. In "emudhra"
// mode it returns the compromised intermediate and its recovered key.
func (c CAConfig) Load() (*tls.Certificate, error) {
	switch strings.ToLower(strings.TrimSpace(c.Mode)) {
	case "", "auto":
		exists, err := c.FilesExist()
		if err != nil {
			return nil, err
		}
		if exists {
			return LoadCertificate(c.CertFile, c.KeyFile)
		}
		return GenerateAndLoad(c.Config, c.CertFile, c.KeyFile, c.IsCA)
	case "emudhra":
		ca, err := c.Resolve()
		if err != nil {
			return nil, err
		}
		return ca.TLSCertificate(c.IncludeRoot)
	default:
		return nil, fmt.Errorf("unknown CA mode %q", c.Mode)
	}
}

// LoadForHost returns a certificate usable as a TLS server certificate for
// host. In "emudhra" mode a leaf certificate is signed with the compromised CA.
func (c CAConfig) LoadForHost(host string, bits int) (*tls.Certificate, error) {
	if !c.IsEmudhra() {
		return c.Load()
	}
	ca, err := c.Resolve()
	if err != nil {
		return nil, err
	}
	if c.Host != "" {
		host = c.Host
	}
	return SignLeafForHost(ca, c.Config, host, bits, c.IncludeRoot)
}

// LoadCertificate loads a certificate and key pair from disk.
func LoadCertificate(certFile string, keyFile string) (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	if cert.Leaf == nil {
		if cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0]); err != nil {
			return nil, err
		}
	}
	return &cert, nil
}

// SignLeafForHost signs a server certificate for host with the given CA.
func SignLeafForHost(ca *emudhra.CA, cfg CertConfig, host string, bits int, includeRoot bool) (*tls.Certificate, error) {
	key, err := ca.Recover()
	if err != nil {
		return nil, err
	}

	// default to 2048 bits, the minimum accepted by clients enforcing
	// OpenSSL security level >= 2
	if bits <= 0 {
		bits = 2048
	}

	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	commonName := host
	if commonName == "" {
		commonName = cfg.CommonName
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:            []string{cfg.Country},
			Locality:           []string{cfg.Locality},
			Organization:       []string{cfg.Organization},
			OrganizationalUnit: []string{cfg.OrganizationalUnit},
			CommonName:         commonName,
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	if ip := net.ParseIP(host); ip != nil {
		template.IPAddresses = []net.IP{ip}
	} else if host != "" {
		template.DNSNames = []string{host}
	}

	der, err := x509.CreateCertificate(rand.Reader, template, ca.Cert, &priv.PublicKey, key)
	if err != nil {
		return nil, err
	}

	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, err
	}

	chain := [][]byte{der, ca.Cert.Raw}
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
		PrivateKey:  priv,
		Leaf:        leaf,
	}, nil
}

func regularFileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("%q is not a regular file", path)
	}
	return true, nil
}
