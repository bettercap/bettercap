package packets

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"sync/atomic"

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
	Sent     uint64 `json:"sent"`
	Received uint64 `json:"received"`
}

type Stats struct {
	Sent        uint64 `json:"sent"`
	Received    uint64 `json:"received"`
	PktReceived uint64 `json:"pkts_received"`
	Errors      uint64 `json:"errors"`
}

type Queue struct {
	sync.RWMutex

	// keep on top because of https://github.com/bettercap/bettercap/issues/500
	Stats      Stats
	Protos     sync.Map
	Traffic    sync.Map
	Activities chan Activity

	iface      *network.Endpoint
	handle     *pcap.Handle
	source     *gopacket.PacketSource
	srcChannel chan gopacket.Packet
	writes     *sync.WaitGroup
	active     bool
}

type queueJSON struct {
	Stats   Stats               `json:"stats"`
	Protos  map[string]int      `json:"protos"`
	Traffic map[string]*Traffic `json:"traffic"`
}

func NewQueue(iface *network.Endpoint) (q *Queue, err error) {
	q = &Queue{
		Protos:     sync.Map{},
		Traffic:    sync.Map{},
		Stats:      Stats{},
		Activities: make(chan Activity),

		writes: &sync.WaitGroup{},
		iface:  iface,
		active: !iface.IsMonitor(),
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

func (q *Queue) MarshalJSON() ([]byte, error) {
	q.Lock()
	defer q.Unlock()
	doc := queueJSON{
		Stats:   q.Stats,
		Protos:  make(map[string]int),
		Traffic: make(map[string]*Traffic),
	}

	q.Protos.Range(func(k, v interface{}) bool {
		doc.Protos[k.(string)] = v.(int)
		return true
	})

	q.Traffic.Range(func(k, v interface{}) bool {
		doc.Traffic[k.(string)] = v.(*Traffic)
		return true
	})

	return json.Marshal(doc)
}

func (q *Queue) trackProtocols(pkt gopacket.Packet) {
	// gather protocols stats
	pktLayers := pkt.Layers()
	for _, layer := range pktLayers {
		proto := layer.LayerType()
		if proto == gopacket.LayerTypeDecodeFailure || proto == gopacket.LayerTypePayload {
			continue
		}

		name := proto.String()
		if v, found := q.Protos.Load(name); !found {
			q.Protos.Store(name, 1)
		} else {
			q.Protos.Store(name, v.(int)+1)
		}
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

	// initialize or update stats
	addr := address.String()
	if v, found := q.Traffic.Load(addr); !found {
		if isSent {
			q.Traffic.Store(addr, &Traffic{Sent: pktSize})
		} else {
			q.Traffic.Store(addr, &Traffic{Received: pktSize})
		}
	} else {
		if isSent {
			v.(*Traffic).Sent += pktSize
		} else {
			v.(*Traffic).Received += pktSize
		}
	}
}

func (q *Queue) TrackPacket(size uint64) {
	// https://github.com/bettercap/bettercap/issues/500
	if q == nil {
		panic("track packet on nil queue!")
	}
	atomic.AddUint64(&q.Stats.PktReceived, 1)
	atomic.AddUint64(&q.Stats.Received, size)
}

func (q *Queue) TrackSent(size uint64) {
	atomic.AddUint64(&q.Stats.Sent, size)
}

func (q *Queue) TrackError() {
	atomic.AddUint64(&q.Stats.Errors, 1)
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
