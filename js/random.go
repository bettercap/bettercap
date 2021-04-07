package js

import (
	"math/rand"
	"net"
	"github.com/bettercap/bettercap/network"
)

type randomPackage struct {
}

func (c randomPackage) String(size int, charset string) string {
	runes := []rune(charset)
	nrunes := len(runes)
	buf := make([]rune, size)
	for i := range buf {
		buf[i] = runes[rand.Intn(nrunes)]
	}
	return string(buf)
}

func (c randomPackage) Mac() string {
	hw := make([]byte, 6)
	rand.Read(hw)
	return network.NormalizeMac(net.HardwareAddr(hw).String())
}