package packets

import (
	"fmt"
	"github.com/google/gopacket/pcap"
	"sync"
)

type Queue struct {
	iface  string
	handle *pcap.Handle
	lock   *sync.Mutex
	active bool
}

func NewQueue(iface string) (*Queue, error) {
	var err error

	q := &Queue{
		iface:  iface,
		handle: nil,
		lock:   &sync.Mutex{},
		active: true,
	}

	q.handle, err = pcap.OpenLive(iface, 1024, true, pcap.BlockForever)
	if err != nil {
		return nil, err
	}

	return q, nil
}

func (q *Queue) Send(raw []byte) error {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.active {
		return q.handle.WritePacketData(raw)
	} else {
		return fmt.Errorf("Packet queue is not active.")
	}
}

func (q *Queue) Stop() {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.handle.Close()
	q.active = false
}
