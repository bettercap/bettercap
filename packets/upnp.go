package packets

import (
	"fmt"
	"net"
)

const (
	UPNPPort = 1900
)

var (
	UPNPDestMac          = net.HardwareAddr{0x01, 0x00, 0x5e, 0x00, 0x00, 0xfb}
	UPNPDestIP           = net.ParseIP("239.255.255.250")
	UPNPDiscoveryPayload = []byte("M-SEARCH * HTTP/1.1\r\n" +
		fmt.Sprintf("Host: %s:%d\r\n", UPNPDestIP, UPNPPort) +
		"Man: ssdp:discover\r\n" +
		"ST: ssdp:all\r\n" +
		"MX: 2\r\n" +
		"\r\n")
)
