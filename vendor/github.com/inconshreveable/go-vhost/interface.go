package vhost

import (
	"net"
)

type Conn interface {
	net.Conn
	Host() string
	Free()
}
