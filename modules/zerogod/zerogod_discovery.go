package zerogod

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/bettercap/bettercap/v2/modules/zerogod/zeroconf"
	"github.com/bettercap/bettercap/v2/network"
	"github.com/bettercap/bettercap/v2/session"
	"github.com/evilsocket/islazy/ops"
	"github.com/evilsocket/islazy/tui"
	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
)

// a service has been discovered
type ServiceDiscoveryEvent struct {
	Service  zeroconf.ServiceEntry `json:"service"`
	Endpoint *network.Endpoint     `json:"endpoint"`
}

// an endpoint is browsing for specific services
type BrowsingEvent struct {
	Source    string            `json:"source"`
	Endpoint  *network.Endpoint `json:"endpoint"`
	Services  []string          `json:"services"`
	Instances []string          `json:"instances"`
	Text      []string          `json:"text"`
	Query     layers.DNS        `json:"query"`
}

func (mod *ZeroGod) onServiceDiscovered(svc *zeroconf.ServiceEntry) {
	mod.Debug("%++v", *svc)

	if svc.Service == DNSSD_DISCOVERY_SERVICE && len(svc.AddrIPv4) == 0 && len(svc.AddrIPv6) == 0 {
		svcName := strings.Replace(svc.Instance, ".local", "", 1)
		if !mod.browser.HasResolverFor(svcName) {
			mod.Debug("discovered service %s", tui.Green(svcName))
			if ch, err := mod.browser.StartBrowsing(svcName, "local.", mod); err != nil {
				mod.Error("%v", err)
			} else {
				// start listening on this channel
				go func() {
					for entry := range ch {
						mod.onServiceDiscovered(entry)
					}
				}()
			}
		}
		return
	}

	mod.Debug("discovered instance %s (%s) [%v / %v]:%d",
		tui.Green(svc.ServiceInstanceName()),
		tui.Dim(svc.HostName),
		svc.AddrIPv4,
		svc.AddrIPv6,
		svc.Port)

	event := ServiceDiscoveryEvent{
		Service:  *svc,
		Endpoint: nil,
	}

	addresses := append(svc.AddrIPv4, svc.AddrIPv6...)

	for _, ip := range addresses {
		address := ip.String()
		if event.Endpoint = mod.Session.Lan.GetByIp(address); event.Endpoint != nil {
			// update internal mapping
			mod.browser.AddServiceFor(address, svc)
			// update endpoint metadata
			mod.updateEndpointMeta(address, event.Endpoint, svc)
			break
		}
	}

	if event.Endpoint == nil {
		mod.Debug("got mdns entry for unknown ip: %++v", *svc)
	}

	session.I.Events.Add("zeroconf.service", event)
	session.I.Refresh()
}

func (mod *ZeroGod) DNSResourceRecord2String(rr *layers.DNSResourceRecord) string {

	if rr.Type == layers.DNSTypeOPT {
		opts := make([]string, len(rr.OPT))
		for i, opt := range rr.OPT {
			opts[i] = opt.String()
		}
		return "OPT " + strings.Join(opts, ",")
	}
	if rr.Type == layers.DNSTypeURI {
		return fmt.Sprintf("URI %d %d %s", rr.URI.Priority, rr.URI.Weight, string(rr.URI.Target))
	}
	/*
	   https://www.rfc-editor.org/rfc/rfc6762

	   Note that the cache-flush bit is NOT part of the resource record
	   class.  The cache-flush bit is the most significant bit of the second
	   16-bit word of a resource record in a Resource Record Section of a
	   Multicast DNS message (the field conventionally referred to as the
	   rrclass field), and the actual resource record class is the least
	   significant fifteen bits of this field.  There is no Multicast DNS
	   resource record class 0x8001.  The value 0x8001 in the rrclass field
	   of a resource record in a Multicast DNS response message indicates a
	   resource record with class 1, with the cache-flush bit set.  When
	   receiving a resource record with the cache-flush bit set,
	   implementations should take care to mask off that bit before storing
	   the resource record in memory, or otherwise ensure that it is given
	   the correct semantic interpretation.
	*/

	if rr.Class == layers.DNSClassIN || rr.Class == 0x8001 {
		switch rr.Type {
		case layers.DNSTypeA, layers.DNSTypeAAAA:
			return rr.IP.String()
		case layers.DNSTypeNS:
			return "NS " + string(rr.NS)
		case layers.DNSTypeCNAME:
			return "CNAME " + string(rr.CNAME)
		case layers.DNSTypePTR:
			return "PTR " + string(rr.PTR)
		case layers.DNSTypeTXT:
			return "TXT \n" + Dump(rr.TXT)
		case layers.DNSTypeSRV:
			return fmt.Sprintf("SRV priority=%d weight=%d port=%d name=%s",
				rr.SRV.Priority,
				rr.SRV.Weight,
				rr.SRV.Port,
				string(rr.SRV.Name))
		case 47: // NSEC
			return "NSEC"
		}
	}

	return fmt.Sprintf("<%v (%d), %v (%d)>", rr.Class, rr.Class, rr.Type, rr.Type)
}

func (mod *ZeroGod) logDNS(src net.IP, dns layers.DNS, isLocal bool) {
	source := tui.Yellow(src.String())
	if endpoint := mod.Session.Lan.GetByIp(src.String()); endpoint != nil {
		if endpoint.Alias != "" {
			source = tui.Bold(endpoint.Alias)
		} else if endpoint.Hostname != "" {
			source = tui.Bold(endpoint.Hostname)
		} else if endpoint.Vendor != "" {
			source = fmt.Sprintf("%s (%s)", tui.Bold(endpoint.IpAddress), tui.Dim(endpoint.Vendor))
		}
	}

	desc := fmt.Sprintf("DNS op=%s %s from %s (r_code=%s)",
		dns.OpCode.String(),
		tui.Bold(ops.Ternary(dns.QR, "RESPONSE", "QUERY").(string)),
		source,
		dns.ResponseCode.String())

	attrs := []string{}
	if dns.AA {
		attrs = append(attrs, "AA")
	}
	if dns.TC {
		attrs = append(attrs, "TC")
	}
	if dns.RD {
		attrs = append(attrs, "RD")
	}
	if dns.RA {
		attrs = append(attrs, "RA")
	}
	if len(attrs) > 0 {
		desc += " [" + strings.Join(attrs, ", ") + "]"
	}

	desc += " :\n"

	for _, q := range dns.Questions {
		desc += fmt.Sprintf("  Q: %s\n", q)
	}
	for _, a := range dns.Answers {
		desc += fmt.Sprintf("  A: %s\n", mod.DNSResourceRecord2String(&a))
	}
	for _, a := range dns.Authorities {
		desc += fmt.Sprintf(" AU: %s\n", mod.DNSResourceRecord2String(&a))
	}
	for _, a := range dns.Additionals {
		desc += fmt.Sprintf(" AD: %s\n", mod.DNSResourceRecord2String(&a))
	}

	if isLocal {
		desc = tui.Dim(desc)
	}

	mod.Info("%s", desc)
}

func (mod *ZeroGod) onPacket(pkt gopacket.Packet) {
	mod.Debug("%++v", pkt)

	// sadly the latest available version of gopacket has an unpatched bug :/
	// https://github.com/bettercap/bettercap/issues/1184
	defer func() {
		if err := recover(); err != nil {
			mod.Error("unexpected error while parsing network packet: %v\n\n%++v", err, pkt)
		}
	}()

	netLayer := pkt.NetworkLayer()
	if netLayer == nil {
		mod.Warning("not network layer in packet %+v", pkt)
		return
	}
	var srcIP net.IP
	// var dstIP net.IP
	switch netLayer.LayerType() {
	case layers.LayerTypeIPv4:
		ip := netLayer.(*layers.IPv4)
		srcIP = ip.SrcIP
		// dstIP = ip.DstIP
	case layers.LayerTypeIPv6:
		ip := netLayer.(*layers.IPv6)
		srcIP = ip.SrcIP
		// dstIP = ip.DstIP
	default:
		mod.Warning("unexpected network layer type %v in packet %+v", netLayer.LayerType(), pkt)
		return
	}

	udp := pkt.Layer(layers.LayerTypeUDP)
	if udp == nil {
		mod.Warning("not udp layer in packet %+v", pkt)
		return
	}

	dns := layers.DNS{}
	if err := dns.DecodeFromBytes(udp.LayerPayload(), gopacket.NilDecodeFeedback); err != nil {
		mod.Warning("could not decode DNS (%v) in packet %+v", err, pkt)
		return
	}

	isLocal := srcIP.Equal(mod.Session.Interface.IP) || srcIP.Equal(mod.Session.Interface.IPv6)

	if _, verbose := mod.BoolParam("zerogod.verbose"); verbose {
		mod.logDNS(srcIP, dns, isLocal)
	}

	// not interested in packet generated by us
	if isLocal {
		mod.Debug("skipping local packet")
		return
	}

	// since the browser is already checking for these, we are only interested in queries
	numQs := len(dns.Questions)
	if numQs == 0 {
		mod.Debug("skipping answers only packet")
		return
	}

	services := make([]string, 0)
	for _, q := range dns.Questions {
		services = append(services, string(q.Name))
	}

	instances := make([]string, 0)
	text := make([]string, 0)
	for _, answer := range append(append(dns.Answers, dns.Additionals...), dns.Authorities...) {
		if answer.Class == layers.DNSClassIN && answer.Type == layers.DNSTypePTR {
			instances = append(instances, string(answer.PTR))
		} else if answer.Type == layers.DNSTypeTXT {
			text = append(text, string(answer.TXT))
		}
	}

	event := BrowsingEvent{
		Source:    srcIP.String(),
		Query:     dns,
		Services:  services,
		Instances: instances,
		Text:      text,
		Endpoint:  mod.Session.Lan.GetByIp(srcIP.String()),
	}

	if event.Endpoint == nil {
		mod.Info("got mdns packet from unknown ip %s", srcIP)
		mod.logDNS(srcIP, dns, isLocal)
		return
	}

	session.I.Events.Add("zeroconf.browsing", event)
	session.I.Refresh()
}

func (mod *ZeroGod) startDiscovery(service string) (err error) {
	mod.Debug("starting resolver for service %s", tui.Yellow(service))

	// create passive sniffer
	if mod.sniffer != nil {
		mod.sniffer.Close()
	}

	readTimeout := 500 * time.Millisecond
	if mod.sniffer, err = network.CaptureWithTimeout(mod.Session.Interface.Name(), readTimeout); err != nil {
		return err
	} else if err = mod.sniffer.SetBPFFilter("udp and port 5353"); err != nil {
		return err
	}
	// prepare source and start listening for packets
	src := gopacket.NewPacketSource(mod.sniffer, mod.sniffer.LinkType())
	mod.snifferCh = src.Packets()
	// start listening for new packets
	go func() {
		mod.Debug("sniffer started")
		for pkt := range mod.snifferCh {
			if !mod.Running() {
				mod.Debug("end pkt loop (pkt=%v)", pkt)
				break
			}
			mod.onPacket(pkt)
		}
		mod.Debug("sniffer stopped")
	}()

	// create service browser
	if mod.browser != nil {
		mod.browser.Stop(false)
	}
	mod.browser = NewBrowser()
	// start active browsing
	if ch, err := mod.browser.StartBrowsing(service, "local.", mod); err != nil {
		return err
	} else {
		// start listening for new services
		go func() {
			for entry := range ch {
				mod.onServiceDiscovered(entry)
			}
		}()
	}

	return nil
}

func (mod *ZeroGod) stopDiscovery() {
	if mod.browser != nil {
		mod.Debug("stopping discovery")
		mod.browser.Stop(true)
		mod.browser = nil
		mod.Debug("discovery stopped")
	}

	if mod.sniffer != nil {
		mod.Debug("stopping sniffer")
		mod.snifferCh <- nil
		mod.sniffer.Close()
		mod.sniffer = nil
		mod.snifferCh = nil
		mod.Debug("sniffer stopped")
	}
}
