package net_sniff

import (
	"crypto/tls"
	"net"
	"testing"
)

// clientHello returns the first TLS record a Go client sends for serverName.
func clientHello(t *testing.T, serverName string) []byte {
	t.Helper()

	client, server := net.Pipe()
	defer server.Close()

	// the handshake fails once the pipe is closed, only the hello is needed
	go func() {
		tls.Client(client, &tls.Config{ServerName: serverName}).Handshake()
		client.Close()
	}()

	buf := make([]byte, 4096)
	n, err := server.Read(buf)
	if err != nil {
		t.Fatalf("could not read client hello: %v", err)
	}
	return buf[:n]
}

func TestSniRe(t *testing.T) {
	// github.com, bbc.com and x.com put a 0x0a byte in one of the SNI
	// extension length fields
	for _, name := range []string{"example.org", "github.com", "bbc.com", "x.com", "www.google.com"} {
		m := sniRe.FindSubmatch(clientHello(t, name))
		if len(m) < 2 {
			t.Errorf("%s: SNI not found", name)
		} else if got := string(m[1]); got != name {
			t.Errorf("%s: got %q", name, got)
		}
	}
}
