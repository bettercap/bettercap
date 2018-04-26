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
	"time"

	"github.com/bettercap/bettercap/log"
)

func getServerCertificate(host string, port int) *x509.Certificate {
	log.Debug("Fetching TLS certificate from %s:%d ...", host, port)

	config := tls.Config{InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", host, port), &config)
	if err != nil {
		log.Warning("Could not fetch TLS certificate from %s:%d: %s", host, port, err)
		return nil
	}
	defer conn.Close()

	state := conn.ConnectionState()

	return state.PeerCertificates[0]
}

func SignCertificateForHost(ca *tls.Certificate, host string, port int) (cert *tls.Certificate, err error) {
	var x509ca *x509.Certificate
	var template x509.Certificate

	if x509ca, err = x509.ParseCertificate(ca.Certificate[0]); err != nil {
		return
	}

	srvCert := getServerCertificate(host, port)
	if srvCert == nil {
		log.Debug("Could not fetch TLS certificate, falling back to default template.")

		notBefore := time.Now()
		aYear := time.Duration(365*24) * time.Hour
		notAfter := notBefore.Add(aYear)
		serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
		serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
		if err != nil {
			return nil, err
		}

		template = x509.Certificate{
			SerialNumber: serialNumber,
			Issuer:       x509ca.Subject,
			Subject: pkix.Name{
				Country:            []string{"US"},
				Locality:           []string{"Scottsdale"},
				Organization:       []string{"GoDaddy.com, Inc."},
				OrganizationalUnit: []string{"https://certs.godaddy.com/repository/"},
				CommonName:         "Go Daddy Secure Certificate Authority - G2",
			},
			NotBefore:             notBefore,
			NotAfter:              notAfter,
			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
		}

		if ip := net.ParseIP(host); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, host)
		}

	} else {
		template = x509.Certificate{
			SerialNumber:          srvCert.SerialNumber,
			Issuer:                x509ca.Subject,
			Subject:               srvCert.Subject,
			NotBefore:             srvCert.NotBefore,
			NotAfter:              srvCert.NotAfter,
			KeyUsage:              srvCert.KeyUsage,
			ExtKeyUsage:           srvCert.ExtKeyUsage,
			IPAddresses:           srvCert.IPAddresses,
			DNSNames:              srvCert.DNSNames,
			BasicConstraintsValid: true,
		}
	}

	var certpriv *rsa.PrivateKey
	if certpriv, err = rsa.GenerateKey(rand.Reader, 1024); err != nil {
		return
	}

	var derBytes []byte
	if derBytes, err = x509.CreateCertificate(rand.Reader, &template, x509ca, &certpriv.PublicKey, ca.PrivateKey); err != nil {
		return
	}

	return &tls.Certificate{
		Certificate: [][]byte{derBytes, ca.Certificate[0]},
		PrivateKey:  certpriv,
	}, nil
}
