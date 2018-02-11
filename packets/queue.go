package packets

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	bnet "github.com/evilsocket/bettercap-ng/net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type Activity struct {
	IP     net.IP
	MAC    net.HardwareAddr
	Source bool
}

type Traffic struct {
	Sent     uint64
	Received uint64
}

type Queue struct {
	sync.Mutex

	Sent        uint64
	Received    uint64
	PktReceived uint64
	Errors      uint64

	Activities chan Activity `json:"-"`
	Protos     map[string]uint64
	Traffic    map[string]*Traffic

	iface  *bnet.Endpoint
	handle *pcap.Handle
	source *gopacket.PacketSource
	active bool
}

func NewQueue(iface *bnet.Endpoint) (*Queue, error) {
	var err error

	q := &Queue{
		iface:       iface,
		handle:      nil,
		active:      true,
		source:      nil,
		Sent:        0,
		Received:    0,
		PktReceived: 0,
		Errors:      0,
		Protos:      make(map[string]uint64),
		Traffic:     make(map[string]*Traffic),
		Activities:  make(chan Activity),
	}

	q.handle, err = pcap.OpenLive(iface.Name(), 1024, true, pcap.BlockForever)
	if err != nil {
		return nil, err
	}

	q.source = gopacket.NewPacketSource(q.handle, q.handle.LinkType())
	go q.worker()

	return q, nil
}

func (q *Queue) trackProtocols(pkt gopacket.Packet) {
	// gather protocols stats
	pktLayers := pkt.Layers()
	for _, layer := range pktLayers {
		proto := layer.LayerType().String()
		if proto == "DecodeFailure" || proto == "Payload" || proto == "Ethernet" {
			continue
		}

		q.Lock()
		if _, found := q.Protos[proto]; found == false {
			q.Protos[proto] = 1
		} else {
			q.Protos[proto] += 1
		}
		q.Unlock()
	}
}

func (q *Queue) trackActivity(eth *layers.Ethernet, ip4 *layers.IPv4, address net.IP, pktSize uint64) {
	// detrmine direction
	isSent := bytes.Compare(address, ip4.SrcIP) == 0
	// push to activity channel
	q.Activities <- Activity{
		IP:     address,
		MAC:    eth.SrcMAC,
		Source: isSent,
	}

	q.Lock()
	defer q.Unlock()

	// initialize or update stats
	addr := address.String()
	if _, found := q.Traffic[addr]; found == false {
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

func (q *Queue) worker() {
	for pkt := range q.source.Packets() {
		if q.active == false {
			return
		}

		q.trackProtocols(pkt)

		pktSize := uint64(len(pkt.Data()))

		atomic.AddUint64(&q.PktReceived, 1)
		atomic.AddUint64(&q.Received, pktSize)

		// decode eth and ipv4 layers
		leth := pkt.Layer(layers.LayerTypeEthernet)
		lip4 := pkt.Layer(layers.LayerTypeIPv4)
		if leth != nil && lip4 != nil {
			eth := leth.(*layers.Ethernet)
			ip4 := lip4.(*layers.IPv4)

			// coming from our network
			if bytes.Compare(q.iface.IP, ip4.SrcIP) != 0 && q.iface.Net.Contains(ip4.SrcIP) {
				q.trackActivity(eth, ip4, ip4.SrcIP, pktSize)
			}
			// coming to our network
			if bytes.Compare(q.iface.IP, ip4.DstIP) != 0 && q.iface.Net.Contains(ip4.DstIP) {
				q.trackActivity(eth, ip4, ip4.DstIP, pktSize)
			}
		}
	}
}

func (q *Queue) Send(raw []byte) error {
	q.Lock()
	defer q.Unlock()

	if q.active == false {
		return fmt.Errorf("Packet queue is not active.")
	}

	if err := q.handle.WritePacketData(raw); err != nil {
		q.Errors += 1
		return err
	} else {
		q.Sent += uint64(len(raw))
	}

	return nil
}

func (q *Queue) Stop() {
	q.Lock()
	defer q.Unlock()
	q.handle.Close()
	q.active = false
}
