package modules

import (
	"github.com/evilsocket/bettercap-ng/net"
)

type ByAddressSorter []*net.Endpoint

func (a ByAddressSorter) Len() int           { return len(a) }
func (a ByAddressSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAddressSorter) Less(i, j int) bool { return a[i].IpAddressUint32 < a[j].IpAddressUint32 }
