package packets

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"

	"github.com/bettercap/bettercap/network"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type Activity struct {
	IP       net.IP
	MAC      net.HardwareAddr
	Hostname string
	Source   bool
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

func (q *Queue) trackActivity(eth *layers.Ethernet, ip4 *layers.IPv4, address net.IP, hostname string, pktSize uint64, isSent bool) {
	// push to activity channel
	q.Activities <- Activity{
		IP:       address,
		MAC:      eth.SrcMAC,
		Hostname: hostname,
		Source:   isSent,
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

func (q *Queue) getHostname(eth *layers.Ethernet, ip *layers.IPv4, pkt gopacket.Packet) string {
	if ludp := pkt.Layer(layers.LayerTypeUDP); ludp != nil {
		if udp := ludp.(*layers.UDP); udp != nil && udp.SrcPort == 5353 && udp.DstPort == 5353 {
			data := udp.Payload
			dataSize := len(data)
			// mDNS query response, no errors
			if dataSize > 4 && data[2] == 0x84 && data[3] == 0x00 {
				// no questions
				if dataSize > 6 && data[4] == 0x00 && data[5] == 0x00 {
					if dataSize > 8 {
						nAnswers := binary.BigEndian.Uint16(data[6:8])
						auth := binary.BigEndian.Uint16(data[9:11])
						addt := binary.BigEndian.Uint16(data[12:14])

						for i := 0; i < nAnswers; i++ {

						}

						return fmt.Sprintf("%d answs", nAnswers)
					}
				}
				return "OK"
			}
		}
	}

	return ""
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
				q.trackActivity(eth, ip4, ip4.SrcIP, q.getHostname(eth, ip4, pkt), pktSize, true)
			}

			// something going to someone on the LAN
			isToMe := q.iface.IP.Equal(ip4.DstIP)
			isToLAN := q.iface.Net.Contains(ip4.DstIP)
			if !isToMe && isToLAN {
				q.trackActivity(eth, ip4, ip4.DstIP, "", pktSize, false)
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
