package vhost

import (
	"bytes"
	"io"
	"net"
	"reflect"
	"testing"
)

func TestHeaderPreserved(t *testing.T) {
	var msg string = "TestHeaderPreserved message! Hello world!"
	var headerLen int = 15

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
		if _, err := conn.Write([]byte(msg)); err != nil {
			panic(err)
		}
		if err = conn.Close(); err != nil {
			panic(err)
		}
	}()

	conn, err := l.Accept()
	if err != nil {
		panic(err)
	}

	// create a shared connection object
	c, rd := newShared(conn)

	// read out a "header"
	p := make([]byte, headerLen)
	_, err = io.ReadFull(rd, p)
	if err != nil {
		panic(err)
	}

	// make sure we got the header
	expectedHeader := []byte(msg[:headerLen])
	if !reflect.DeepEqual(p, expectedHeader) {
		t.Errorf("Read header bytes %s, expected %s", p, expectedHeader)
		return
	}

	// read out the entire connection. make sure it includes the header
	buf := bytes.NewBuffer([]byte{})
	io.Copy(buf, c)

	expected := []byte(msg)
	if !reflect.DeepEqual(buf.Bytes(), expected) {
		t.Errorf("Read full connection bytes %s, expected %s", buf.Bytes(), expected)
	}
}
