package packets

import (
	"fmt"
	"github.com/google/gopacket/pcap"
	"github.com/op/go-logging"
	"sync"
)

var log = logging.MustGetLogger("mitm")

type Queue struct {
	iface  string
	handle *pcap.Handle
	lock   *sync.Mutex
	active bool
}

func NewQueue(iface string) (*Queue, error) {
	log.Debugf("Creating packet queue for interface %s.\n", iface)
	var err error

	q := &Queue{
		iface:  iface,
		handle: nil,
		lock:   &sync.Mutex{},
		active: true,
	}

	q.handle, err = pcap.OpenLive(iface, 65536, true, pcap.BlockForever)
	if err != nil {
		return nil, err
	}

	return q, nil
}

func (q *Queue) Send(raw []byte) error {
	q.lock.Lock()
	defer q.lock.Unlock()

	log.Debugf("Sending %d bytes to packet queue.\n", len(raw))
	if q.active {
		return q.handle.WritePacketData(raw)
	} else {
		return fmt.Errorf("Packet queue is not active.")
	}
}

func (q *Queue) Stop() {
	q.lock.Lock()
	defer q.lock.Unlock()

	log.Debugf("Stopping packet queue.\n")

	q.handle.Close()
	q.active = false
}
