package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math/big"
	"sort"
)

func hashSorted(lst []string) []byte {
	c := make([]string, len(lst))
	copy(c, lst)
	sort.Strings(c)
	h := sha1.New()
	for _, s := range c {
		h.Write([]byte(s + ","))
	}
	return h.Sum(nil)
}

func hashSortedBigInt(lst []string) *big.Int {
	rv := new(big.Int)
	rv.SetBytes(hashSorted(lst))
	return rv
}

func getServerCertificate(host string, port int) *x509.Certificate {
	config := tls.Config{InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", host, port), &config)
	if err != nil {
		return nil
	}
	defer conn.Close()

	state := conn.ConnectionState()

	return state.PeerCertificates[0]
}

func SignCertificateForHost(ca *tls.Certificate, host string, port int) (cert *tls.Certificate, err error) {
	var x509ca *x509.Certificate

	if x509ca, err = x509.ParseCertificate(ca.Certificate[0]); err != nil {
		return
	}

	srvCert := getServerCertificate(host, port)

	template := x509.Certificate{
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
