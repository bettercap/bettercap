package packets

import (
	"fmt"
	"net"
	"sync"

	"github.com/bettercap/bettercap/network"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type Activity struct {
	IP     net.IP
	MAC    net.HardwareAddr
	Meta   map[string]string
	Source bool
}

type Traffic struct {
	Sent     uint64
	Received uint64
}

type Stats struct {
	sync.RWMutex

	Sent        uint64
	Received    uint64
	PktReceived uint64
	Errors      uint64
}

type PacketCallback func(pkt gopacket.Packet)

type Queue struct {
	sync.RWMutex

	Activities chan Activity `json:"-"`

	Stats   Stats
	Protos  map[string]uint64
	Traffic map[string]*Traffic

	iface      *network.Endpoint
	handle     *pcap.Handle
	source     *gopacket.PacketSource
	srcChannel chan gopacket.Packet
	writes     *sync.WaitGroup
	pktCb      PacketCallback
	active     bool
}

func NewQueue(iface *network.Endpoint) (q *Queue, err error) {
	q = &Queue{
		Protos:     make(map[string]uint64),
		Traffic:    make(map[string]*Traffic),
		Activities: make(chan Activity),

		writes: &sync.WaitGroup{},
		iface:  iface,
		active: !iface.IsMonitor(),
		pktCb:  nil,
	}

	if q.active {
		if q.handle, err = pcap.OpenLive(iface.Name(), 1024, true, pcap.BlockForever); err != nil {
			return
		}

		q.source = gopacket.NewPacketSource(q.handle, q.handle.LinkType())
		q.srcChannel = q.source.Packets()
		go q.worker()
	}

	return
}

func (q *Queue) OnPacket(cb PacketCallback) {
	q.Lock()
	defer q.Unlock()
	q.pktCb = cb
}

func (q *Queue) onPacketCallback(pkt gopacket.Packet) {
	q.RLock()
	defer q.RUnlock()

	if q.pktCb != nil {
		q.pktCb(pkt)
	}
}

func (q *Queue) trackProtocols(pkt gopacket.Packet) {
	// gather protocols stats
	pktLayers := pkt.Layers()
	for _, layer := range pktLayers {
		proto := layer.LayerType()
		if proto == gopacket.LayerTypeDecodeFailure || proto == gopacket.LayerTypePayload {
			continue
		}

		q.Lock()
		name := proto.String()
		if _, found := q.Protos[name]; !found {
			q.Protos[name] = 1
		} else {
			q.Protos[name]++
		}
		q.Unlock()
	}
}

func (q *Queue) trackActivity(eth *layers.Ethernet, ip4 *layers.IPv4, address net.IP, meta map[string]string, pktSize uint64, isSent bool) {
	// push to activity channel
	q.Activities <- Activity{
		IP:     address,
		MAC:    eth.SrcMAC,
		Meta:   meta,
		Source: isSent,
	}

	q.Lock()
	defer q.Unlock()

	// initialize or update stats
	addr := address.String()
	if _, found := q.Traffic[addr]; !found {
		if isSent {
			q.Traffic[addr] = &Traffic{Sent: pktSize}
		} else {
			q.Traffic[addr] = &Traffic{Received: pktSize}
		}
	} else {
		if isSent {
			q.Traffic[addr].Sent += pktSize
		} else {
			q.Traffic[addr].Received += pktSize
		}
	}
}

func (q *Queue) TrackPacket(size uint64) {
	q.Stats.Lock()
	defer q.Stats.Unlock()

	q.Stats.PktReceived++
	q.Stats.Received += size
}

func (q *Queue) TrackSent(size uint64) {
	q.Stats.Lock()
	defer q.Stats.Unlock()

	q.Stats.Sent += size
}

func (q *Queue) TrackError() {
	q.Stats.Lock()
	defer q.Stats.Unlock()

	q.Stats.Errors++
}

func (q *Queue) getPacketMeta(pkt gopacket.Packet) map[string]string {
	meta := make(map[string]string)
	if mdns := MDNSGetMeta(pkt); mdns != nil {
		meta = mdns
	} else if nbns := NBNSGetMeta(pkt); nbns != nil {
		meta = nbns
	} else if upnp := UPNPGetMeta(pkt); upnp != nil {
		meta = upnp
	}
	return meta
}

func (q *Queue) worker() {
	for pkt := range q.srcChannel {
		if !q.active {
			return
		}

		q.trackProtocols(pkt)

		pktSize := uint64(len(pkt.Data()))

		q.TrackPacket(pktSize)
		q.onPacketCallback(pkt)

		// decode eth and ipv4 layers
		leth := pkt.Layer(layers.LayerTypeEthernet)
		lip4 := pkt.Layer(layers.LayerTypeIPv4)
		if leth != nil && lip4 != nil {
			eth := leth.(*layers.Ethernet)
			ip4 := lip4.(*layers.IPv4)

			// here we try to discover new hosts
			// on this lan by inspecting packets
			// we manage to sniff

			// something coming from someone on the LAN
			isFromMe := q.iface.IP.Equal(ip4.SrcIP)
			isFromLAN := q.iface.Net.Contains(ip4.SrcIP)
			if !isFromMe && isFromLAN {
				meta := q.getPacketMeta(pkt)

				q.trackActivity(eth, ip4, ip4.SrcIP, meta, pktSize, true)
			}

			// something going to someone on the LAN
			isToMe := q.iface.IP.Equal(ip4.DstIP)
			isToLAN := q.iface.Net.Contains(ip4.DstIP)
			if !isToMe && isToLAN {
				q.trackActivity(eth, ip4, ip4.DstIP, nil, pktSize, false)
			}
		}
	}
}

func (q *Queue) Send(raw []byte) error {
	q.Lock()
	defer q.Unlock()

	if !q.active {
		return fmt.Errorf("Packet queue is not active.")
	}

	q.writes.Add(1)
	defer q.writes.Done()

	if err := q.handle.WritePacketData(raw); err != nil {
		q.TrackError()
		return err
	} else {
		q.TrackSent(uint64(len(raw)))
	}

	return nil
}

func (q *Queue) Stop() {
	q.Lock()
	defer q.Unlock()

	if q.active {
		// wait for write operations to be completed
		q.writes.Wait()
		// signal the main loop to exit and close the handle
		q.active = false
		q.srcChannel <- nil
		q.handle.Close()
	}
}
