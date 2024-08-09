package modules

import (
	"github.com/bettercap/bettercap/v2/modules/any_proxy"
	"github.com/bettercap/bettercap/v2/modules/api_rest"
	"github.com/bettercap/bettercap/v2/modules/arp_spoof"
	"github.com/bettercap/bettercap/v2/modules/ble"
	"github.com/bettercap/bettercap/v2/modules/c2"
	"github.com/bettercap/bettercap/v2/modules/can"
	"github.com/bettercap/bettercap/v2/modules/caplets"
	"github.com/bettercap/bettercap/v2/modules/dhcp6_spoof"
	"github.com/bettercap/bettercap/v2/modules/dns_spoof"
	"github.com/bettercap/bettercap/v2/modules/events_stream"
	"github.com/bettercap/bettercap/v2/modules/gps"
	"github.com/bettercap/bettercap/v2/modules/graph"
	"github.com/bettercap/bettercap/v2/modules/hid"
	"github.com/bettercap/bettercap/v2/modules/http_proxy"
	"github.com/bettercap/bettercap/v2/modules/http_server"
	"github.com/bettercap/bettercap/v2/modules/https_proxy"
	"github.com/bettercap/bettercap/v2/modules/https_server"
	"github.com/bettercap/bettercap/v2/modules/mac_changer"
	"github.com/bettercap/bettercap/v2/modules/mdns_server"
	"github.com/bettercap/bettercap/v2/modules/mysql_server"
	"github.com/bettercap/bettercap/v2/modules/ndp_spoof"
	"github.com/bettercap/bettercap/v2/modules/net_probe"
	"github.com/bettercap/bettercap/v2/modules/net_recon"
	"github.com/bettercap/bettercap/v2/modules/net_sniff"
	"github.com/bettercap/bettercap/v2/modules/packet_proxy"
	"github.com/bettercap/bettercap/v2/modules/syn_scan"
	"github.com/bettercap/bettercap/v2/modules/tcp_proxy"
	"github.com/bettercap/bettercap/v2/modules/ticker"
	"github.com/bettercap/bettercap/v2/modules/ui"
	"github.com/bettercap/bettercap/v2/modules/update"
	"github.com/bettercap/bettercap/v2/modules/wifi"
	"github.com/bettercap/bettercap/v2/modules/wol"

	"github.com/bettercap/bettercap/v2/session"
)

func LoadModules(sess *session.Session) {
	sess.Register(any_proxy.NewAnyProxy(sess))
	sess.Register(arp_spoof.NewArpSpoofer(sess))
	sess.Register(api_rest.NewRestAPI(sess))
	sess.Register(ble.NewBLERecon(sess))
	sess.Register(can.NewCanModule(sess))
	sess.Register(dhcp6_spoof.NewDHCP6Spoofer(sess))
	sess.Register(net_recon.NewDiscovery(sess))
	sess.Register(dns_spoof.NewDNSSpoofer(sess))
	sess.Register(events_stream.NewEventsStream(sess))
	sess.Register(gps.NewGPS(sess))
	sess.Register(graph.NewModule(sess))
	sess.Register(http_proxy.NewHttpProxy(sess))
	sess.Register(http_server.NewHttpServer(sess))
	sess.Register(https_proxy.NewHttpsProxy(sess))
	sess.Register(https_server.NewHttpsServer(sess))
	sess.Register(mac_changer.NewMacChanger(sess))
	sess.Register(mysql_server.NewMySQLServer(sess))
	sess.Register(mdns_server.NewMDNSServer(sess))
	sess.Register(net_sniff.NewSniffer(sess))
	sess.Register(packet_proxy.NewPacketProxy(sess))
	sess.Register(net_probe.NewProber(sess))
	sess.Register(syn_scan.NewSynScanner(sess))
	sess.Register(tcp_proxy.NewTcpProxy(sess))
	sess.Register(ticker.NewTicker(sess))
	sess.Register(wifi.NewWiFiModule(sess))
	sess.Register(wol.NewWOL(sess))
	sess.Register(hid.NewHIDRecon(sess))
	sess.Register(c2.NewC2(sess))
	sess.Register(ndp_spoof.NewNDPSpoofer(sess))

	sess.Register(caplets.NewCapletsModule(sess))
	sess.Register(update.NewUpdateModule(sess))
	sess.Register(ui.NewUIModule(sess))
}
