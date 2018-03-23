package iprange

import (
	"math/big"
	"net"
)

// Asc implements sorting in ascending order for IP addresses
type asc []net.IP

func (a asc) Len() int {
	return len(a)
}

func (a asc) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a asc) Less(i, j int) bool {
	bigi := big.NewInt(0).SetBytes(a[i])
	bigj := big.NewInt(0).SetBytes(a[j])

	if bigi.Cmp(bigj) == -1 {
		return true
	}
	return false
}
