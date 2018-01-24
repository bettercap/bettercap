package modules

import (
	"crypto/tls"
	"fmt"
	"sync"
)

var (
	certCache = make(map[string]*tls.Certificate)
	certLock  = &sync.Mutex{}
)

func getCachedCert(domain string, port int) *tls.Certificate {
	key := fmt.Sprintf("%s:%d", domain, port)

	certLock.Lock()
	defer certLock.Unlock()

	if cert, found := certCache[key]; found == true {
		return cert
	}
	return nil
}

func setCachedCert(domain string, port int, cert *tls.Certificate) {
	key := fmt.Sprintf("%s:%d", domain, port)

	certLock.Lock()
	defer certLock.Unlock()

	certCache[key] = cert
}
