package packets

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/evilsocket/islazy/str"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
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

func UPNPGetMeta(pkt gopacket.Packet) map[string]string {
	if ludp := pkt.Layer(layers.LayerTypeUDP); ludp != nil {
		if udp := ludp.(*layers.UDP); udp != nil && udp.SrcPort == UPNPPort && len(udp.Payload) > 0 {
			request := &http.Request{}
			reader := bufio.NewReader(bytes.NewReader(udp.Payload))
			if response, err := http.ReadResponse(reader, request); err == nil {
				meta := make(map[string]string)
				for name, values := range response.Header {
					if name != "Cache-Control" && len(values) > 0 {
						if data := str.Trim(strings.Join(values, ", ")); data != "" {
							meta["upnp:"+name] = data
						}

					}
				}
				return meta
			}
		}
	}
	return nil
}
