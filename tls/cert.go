package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/bettercap/bettercap/session"
)

type CertConfig struct {
	Bits               int
	Country            string
	Locality           string
	Organization       string
	OrganizationalUnit string
	CommonName         string
}

var (
	DefaultLegitConfig = CertConfig{
		Bits:               4096,
		Country:            "US",
		Locality:           "",
		Organization:       "bettercap devteam",
		OrganizationalUnit: "https://bettercap.org/",
		CommonName:         "bettercap",
	}
	DefaultSpoofConfig = CertConfig{
		Bits:               4096,
		Country:            "US",
		Locality:           "Scottsdale",
		Organization:       "GoDaddy.com, Inc.",
		OrganizationalUnit: "https://certs.godaddy.com/repository/",
		CommonName:         "Go Daddy Secure Certificate Authority - G2",
	}
)

func CertConfigToModule(prefix string, m *session.SessionModule, defaults CertConfig) {
	m.AddParam(session.NewIntParameter(prefix+".certificate.bits", strconv.Itoa(defaults.Bits),
		"Number of bits of the RSA private key of the generated HTTPS certificate."))
	m.AddParam(session.NewStringParameter(prefix+".certificate.country", defaults.Country, ".*",
		"Country field of the generated HTTPS certificate."))
	m.AddParam(session.NewStringParameter(prefix+".certificate.locality", defaults.Locality, ".*",
		"Locality field of the generated HTTPS certificate."))
	m.AddParam(session.NewStringParameter(prefix+".certificate.organization", defaults.Organization, ".*",
		"Organization field of the generated HTTPS certificate."))
	m.AddParam(session.NewStringParameter(prefix+".certificate.organizationalunit", defaults.OrganizationalUnit, ".*",
		"Organizational Unit field of the generated HTTPS certificate."))
	m.AddParam(session.NewStringParameter(prefix+".certificate.commonname", defaults.CommonName, ".*",
		"Common Name field of the generated HTTPS certificate."))
}

func CertConfigFromModule(prefix string, m session.SessionModule) (cfg CertConfig, err error) {
	if err, cfg.Bits = m.IntParam(prefix + ".certificate.bits"); err != nil {
		return cfg, err
	} else if err, cfg.Country = m.StringParam(prefix + ".certificate.country"); err != nil {
		return cfg, err
	} else if err, cfg.Locality = m.StringParam(prefix + ".certificate.locality"); err != nil {
		return cfg, err
	} else if err, cfg.Organization = m.StringParam(prefix + ".certificate.organization"); err != nil {
		return cfg, err
	} else if err, cfg.OrganizationalUnit = m.StringParam(prefix + ".certificate.organizationalunit"); err != nil {
		return cfg, err
	} else if err, cfg.CommonName = m.StringParam(prefix + ".certificate.commonname"); err != nil {
		return cfg, err
	}
	return cfg, err
}

func CreateCertificate(cfg CertConfig, ca bool) (*rsa.PrivateKey, []byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, cfg.Bits)
	if err != nil {
		return nil, nil, err
	}

	notBefore := time.Now()
	aYear := time.Duration(365*24) * time.Hour
	notAfter := notBefore.Add(aYear)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:            []string{cfg.Country},
			Locality:           []string{cfg.Locality},
			Organization:       []string{cfg.Organization},
			OrganizationalUnit: []string{cfg.OrganizationalUnit},
			CommonName:         cfg.CommonName,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  ca,
	}

	cert, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	return priv, cert, err
}

func Generate(cfg CertConfig, certPath string, keyPath string, ca bool) error {
	keyFile, err := os.Create(keyPath)
	if err != nil {
		return err
	}
	defer keyFile.Close()

	certFile, err := os.Create(certPath)
	if err != nil {
		return err
	}
	defer certFile.Close()

	priv, cert, err := CreateCertificate(cfg, ca)
	if err != nil {
		return err
	}

	if err := pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}); err != nil {
		return err
	}

	return pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: cert})
}
