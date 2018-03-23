package vhost

import (
	"net"
	"net/http"
	"testing"
)

func TestHTTPHost(t *testing.T) {
	var testHostname string = "foo.example.com"

	l, err := net.Listen("tcp", "127.0.0.1:12345")
	if err != nil {
		panic(err)
	}
	defer l.Close()

	go func() {
		conn, err := net.Dial("tcp", "127.0.0.1:12345")
		if err != nil {
			panic(err)
		}
		defer conn.Close()
		req, err := http.NewRequest("GET", "http://"+testHostname+"/bar", nil)
		if err != nil {
			panic(err)
		}
		if err = req.Write(conn); err != nil {
			panic(err)
		}
	}()

	conn, err := l.Accept()
	if err != nil {
		panic(err)
	}
	c, err := HTTP(conn)
	if err != nil {
		panic(err)
	}

	if c.Host() != testHostname {
		t.Errorf("Connection Host() is %s, expected %s", c.Host(), testHostname)
	}
}
