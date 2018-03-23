package vhost

import (
	"crypto/tls"
	"net"
	"testing"
)

func TestSNI(t *testing.T) {
	var testHostname string = "foo.example.com"

	l, err := net.Listen("tcp", "127.0.0.1:12345")
	if err != nil {
		panic(err)
	}
	defer l.Close()

	go func() {
		conf := &tls.Config{ServerName: testHostname}
		conn, err := tls.Dial("tcp", "127.0.0.1:12345", conf)
		if err != nil {
			panic(err)
		}
		conn.Close()
	}()

	conn, err := l.Accept()
	if err != nil {
		panic(err)
	}
	c, err := TLS(conn)
	if err != nil {
		panic(err)
	}

	if c.Host() != testHostname {
		t.Errorf("Connection Host() is %s, expected %s", c.Host(), testHostname)
	}
}
