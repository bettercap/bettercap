package modules

import (
	"github.com/bettercap/bettercap/modules/any_proxy"
	"github.com/bettercap/bettercap/modules/api_rest"
	"github.com/bettercap/bettercap/modules/arp_spoof"
	"github.com/bettercap/bettercap/modules/ble"
	"github.com/bettercap/bettercap/modules/c2"
	"github.com/bettercap/bettercap/modules/caplets"
	"github.com/bettercap/bettercap/modules/dhcp6_spoof"
	"github.com/bettercap/bettercap/modules/dns_spoof"
	"github.com/bettercap/bettercap/modules/events_stream"
	"github.com/bettercap/bettercap/modules/gps"
	"github.com/bettercap/bettercap/modules/hid"
	"github.com/bettercap/bettercap/modules/http_proxy"
	"github.com/bettercap/bettercap/modules/http_server"
	"github.com/bettercap/bettercap/modules/https_proxy"
	"github.com/bettercap/bettercap/modules/https_server"
	"github.com/bettercap/bettercap/modules/mac_changer"
	"github.com/bettercap/bettercap/modules/mdns_server"
	"github.com/bettercap/bettercap/modules/mysql_server"
	"github.com/bettercap/bettercap/modules/ndp_spoof"
	"github.com/bettercap/bettercap/modules/net_probe"
	"github.com/bettercap/bettercap/modules/net_recon"
	"github.com/bettercap/bettercap/modules/net_sniff"
	"github.com/bettercap/bettercap/modules/packet_proxy"
	"github.com/bettercap/bettercap/modules/syn_scan"
	"github.com/bettercap/bettercap/modules/tcp_proxy"
	"github.com/bettercap/bettercap/modules/ticker"
	"github.com/bettercap/bettercap/modules/ui"
	"github.com/bettercap/bettercap/modules/update"
	"github.com/bettercap/bettercap/modules/wifi"
	"github.com/bettercap/bettercap/modules/wol"

	"github.com/bettercap/bettercap/session"
)

func LoadModules(sess *session.Session) {
	sess.Register(any_proxy.NewAnyProxy(sess))
	sess.Register(arp_spoof.NewArpSpoofer(sess))
	sess.Register(api_rest.NewRestAPI(sess))
	sess.Register(ble.NewBLERecon(sess))
	sess.Register(dhcp6_spoof.NewDHCP6Spoofer(sess))
	sess.Register(net_recon.NewDiscovery(sess))
	sess.Register(dns_spoof.NewDNSSpoofer(sess))
	sess.Register(events_stream.NewEventsStream(sess))
	sess.Register(gps.NewGPS(sess))
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
