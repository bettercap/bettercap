package http_proxy

import (
	"crypto/tls"
	"fmt"
	"sync"
)

var (
	certCache = make(map[string]*tls.Certificate)
	certLock  = &sync.Mutex{}
)

func keyFor(domain string, port int) string {
	return fmt.Sprintf("%s:%d", domain, port)
}

func getCachedCert(domain string, port int) *tls.Certificate {
	certLock.Lock()
	defer certLock.Unlock()
	if cert, found := certCache[keyFor(domain, port)]; found {
		return cert
	}
	return nil
}

func setCachedCert(domain string, port int, cert *tls.Certificate) {
	certLock.Lock()
	defer certLock.Unlock()
	certCache[keyFor(domain, port)] = cert
}

// ClearCertCache drops every cached spoofed certificate. It must be called
// whenever the CA changes, since cached certificates are signed by the
// previous CA.
func ClearCertCache() {
	certLock.Lock()
	defer certLock.Unlock()
	certCache = make(map[string]*tls.Certificate)
}
