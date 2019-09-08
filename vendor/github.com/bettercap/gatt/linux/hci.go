package linux

import (
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/bettercap/gatt/linux/cmd"
	"github.com/bettercap/gatt/linux/evt"
	"github.com/bettercap/gatt/linux/util"
)

type HCI struct {
	AcceptMasterHandler  func(pd *PlatData)
	AcceptSlaveHandler   func(pd *PlatData)
	AdvertisementHandler func(pd *PlatData)

	d io.ReadWriteCloser
	c *cmd.Cmd
	e *evt.Evt

	plist   map[bdaddr]*PlatData
	plistmu *sync.Mutex

	bufCnt  chan struct{}
	bufSize int

	pool     *util.BytePool
	loopDone chan bool

	maxConn int
	connsmu *sync.Mutex
	conns   map[uint16]*conn

	adv   bool
	advmu *sync.Mutex
}

type bdaddr [6]byte

type PlatData struct {
	Name        string
	AddressType uint8
	Address     [6]byte
	Data        []byte
	Connectable bool
	RSSI        int8

	Conn io.ReadWriteCloser
}

func NewHCI(devID int, chk bool, maxConn int) (*HCI, error) {
	d, err := newDevice(devID, chk)
	if err != nil {
		return nil, err
	}
	c := cmd.NewCmd(d)
	e := evt.NewEvt()

	h := &HCI{
		d: d,
		c: c,
		e: e,

		plist:   make(map[bdaddr]*PlatData),
		plistmu: &sync.Mutex{},

		bufCnt:  make(chan struct{}, 15-1),
		bufSize: 27,

		pool:     util.NewBytePool(4096, 16),
		loopDone: make(chan bool),

		maxConn: maxConn,
		connsmu: &sync.Mutex{},
		conns:   map[uint16]*conn{},

		advmu: &sync.Mutex{},
	}

	e.HandleEvent(evt.LEMeta, evt.HandlerFunc(h.handleLEMeta))
	e.HandleEvent(evt.DisconnectionComplete, evt.HandlerFunc(h.handleDisconnectionComplete))
	e.HandleEvent(evt.NumberOfCompletedPkts, evt.HandlerFunc(h.handleNumberOfCompletedPkts))
	e.HandleEvent(evt.CommandComplete, evt.HandlerFunc(c.HandleComplete))
	e.HandleEvent(evt.CommandStatus, evt.HandlerFunc(c.HandleStatus))

	go h.mainLoop()
	h.resetDevice()
	return h, nil
}

func (h *HCI) Close() error {
	log.Printf("hci.Close()")
	h.pool.Put(nil)
	<-h.loopDone
	log.Printf("mainLoop exited")
	for _, c := range h.conns {
		log.Printf("closing connection %v", c)
		c.Close()
	}
	log.Printf("closing %v", h.d)
	return h.d.Close()
}

func (h *HCI) SetAdvertiseEnable(en bool) error {
	h.advmu.Lock()
	h.adv = en
	h.advmu.Unlock()
	return h.setAdvertiseEnable(en)
}

func (h *HCI) setAdvertiseEnable(en bool) error {
	h.advmu.Lock()
	defer h.advmu.Unlock()
	if en && h.adv && (len(h.conns) == h.maxConn) {
		return nil
	}
	return h.c.SendAndCheckResp(
		cmd.LESetAdvertiseEnable{
			AdvertisingEnable: btoi(en),
		}, []byte{0x00})
}

func (h *HCI) SendCmdWithAdvOff(c cmd.CmdParam) error {
	h.setAdvertiseEnable(false)
	err := h.c.SendAndCheckResp(c, nil)
	if h.adv {
		h.setAdvertiseEnable(true)
	}
	return err
}

func (h *HCI) SetScanEnable(en bool, dup bool) error {
	return h.c.SendAndCheckResp(
		cmd.LESetScanEnable{
			LEScanEnable:     btoi(en),
			FilterDuplicates: btoi(!dup),
		}, []byte{0x00})
}

func (h *HCI) Connect(pd *PlatData) error {
	h.c.Send(
		cmd.LECreateConn{
			LEScanInterval:        0x0004,         // N x 0.625ms
			LEScanWindow:          0x0004,         // N x 0.625ms
			InitiatorFilterPolicy: 0x00,           // white list not used
			PeerAddressType:       pd.AddressType, // public or random
			PeerAddress:           pd.Address,     //
			OwnAddressType:        0x00,           // public
			ConnIntervalMin:       0x0006,         // N x 0.125ms
			ConnIntervalMax:       0x0006,         // N x 0.125ms
			ConnLatency:           0x0000,         //
			SupervisionTimeout:    0x0048,         // N x 10ms
			MinimumCELength:       0x0000,         // N x 0.625ms
			MaximumCELength:       0x0000,         // N x 0.625ms
		})
	return nil
}

func (h *HCI) CancelConnection(pd *PlatData) error {
	if pd != nil && pd.Conn != nil {
		return pd.Conn.Close()
	}
	return nil
}

func (h *HCI) SendRawCommand(c cmd.CmdParam) ([]byte, error) {
	return h.c.Send(c)
}

func btoi(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

func (h *HCI) mainLoop() {
	log.Printf("hci.mainLoop started")
	defer func() {
		h.loopDone <- true
	}()

	for {
		// log.Printf("hci.mainLoop pool.Get")
		b := h.pool.Get()
		if b == nil {
			log.Printf("got nil buffer, breaking mainLoop")
			break
		}
		// log.Printf("hci.mainLoop Read(%d)", len(b))
		n, err := h.d.Read(b)
		if err != nil {
			log.Printf("mainloop err: %v", err)
			return
		}
		if n == 0 {
			log.Printf("mainLoop failed to read")
			return
		}

		// log.Printf("hci.mainLoop -> handlePacket")
		h.handlePacket(b, n)
	}
	log.Printf("hci.mainLoop stopped")
}

func (h *HCI) handlePacket(buf []byte, n int) {
	b := buf[:n]
	t, b := packetType(b[0]), b[1:]
	var err error
	handled := true
	switch t {
	case typCommandPkt:
		op := uint16(b[0]) | uint16(b[1])<<8
		log.Printf("unmanaged cmd: opcode (%04x) [ % X ]\n", op, b)
	case typACLDataPkt:
		err = h.handleL2CAP(b)
	case typSCODataPkt:
		err = fmt.Errorf("SCO packet not supported")
	case typEventPkt:
		handled = false
		go func() {
			err := h.e.Dispatch(b)
			if err != nil {
				log.Printf("hci: %s, [ % X]", err, b)
			}
			h.pool.Put(buf)
		}()
	case typVendorPkt:
		err = fmt.Errorf("Vendor packet not supported")
	default:
		log.Fatalf("Unknown event: 0x%02X [ % X ]\n", t, b)
	}
	if err != nil {
		log.Printf("hci: %s, [ % X]", err, b)
	}
	if handled {
		h.pool.Put(buf)
	}
}

func (h *HCI) resetDevice() error {
	seq := []cmd.CmdParam{
		cmd.Reset{},
		cmd.SetEventMask{EventMask: 0x3dbff807fffbffff},
		cmd.LESetEventMask{LEEventMask: 0x000000000000001F},
		cmd.WriteSimplePairingMode{SimplePairingMode: 1},
		cmd.WriteLEHostSupported{LESupportedHost: 1, SimultaneousLEHost: 0},
		cmd.WriteInquiryMode{InquiryMode: 2},
		cmd.WritePageScanType{PageScanType: 1},
		cmd.WriteInquiryScanType{ScanType: 1},
		cmd.WriteClassOfDevice{ClassOfDevice: [3]byte{0x40, 0x02, 0x04}},
		cmd.WritePageTimeout{PageTimeout: 0x2000},
		cmd.WriteDefaultLinkPolicy{DefaultLinkPolicySettings: 0x5},
		cmd.HostBufferSize{
			HostACLDataPacketLength:            0x1000,
			HostSynchronousDataPacketLength:    0xff,
			HostTotalNumACLDataPackets:         0x0014,
			HostTotalNumSynchronousDataPackets: 0x000a},
		cmd.LESetScanParameters{
			LEScanType:           0x01,   // [0x00]: passive, 0x01: active
			LEScanInterval:       0x0010, // [0x10]: 0.625ms * 16
			LEScanWindow:         0x0010, // [0x10]: 0.625ms * 16
			OwnAddressType:       0x00,   // [0x00]: public, 0x01: random
			ScanningFilterPolicy: 0x00,   // [0x00]: accept all, 0x01: ignore non-white-listed.
		},
	}
	for _, s := range seq {
		if err := h.c.SendAndCheckResp(s, []byte{0x00}); err != nil {
			return err
		}
	}
	return nil
}

func (h *HCI) handleAdvertisement(b []byte) {
	// If no one is interested, don't bother.
	if h.AdvertisementHandler == nil {
		return
	}
	ep := &evt.LEAdvertisingReportEP{}
	if err := ep.Unmarshal(b); err != nil {
		return
	}
	for i := 0; i < int(ep.NumReports); i++ {
		addr := bdaddr(ep.Address[i])
		et := ep.EventType[i]
		connectable := et == advInd || et == advDirectInd
		scannable := et == advInd || et == advScanInd

		if et == scanRsp {
			h.plistmu.Lock()
			pd, ok := h.plist[addr]
			h.plistmu.Unlock()
			if ok {
				pd.Data = append(pd.Data, ep.Data[i]...)
				h.AdvertisementHandler(pd)
			}
			continue
		}

		pd := &PlatData{
			AddressType: ep.AddressType[i],
			Address:     ep.Address[i],
			Data:        ep.Data[i],
			Connectable: connectable,
			RSSI:        ep.RSSI[i],
		}
		h.plistmu.Lock()
		h.plist[addr] = pd
		h.plistmu.Unlock()
		if scannable {
			continue
		}
		h.AdvertisementHandler(pd)
	}
}

func (h *HCI) handleNumberOfCompletedPkts(b []byte) error {
	ep := &evt.NumberOfCompletedPktsEP{}
	if err := ep.Unmarshal(b); err != nil {
		return err
	}
	for _, r := range ep.Packets {
		for i := 0; i < int(r.NumOfCompletedPkts); i++ {
			<-h.bufCnt
		}
	}
	return nil
}

func (h *HCI) handleConnection(b []byte) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("error while handling connectiont: %v", err)
		}
	}()

	ep := &evt.LEConnectionCompleteEP{}
	if err := ep.Unmarshal(b); err != nil {
		return // FIXME
	}
	hh := ep.ConnectionHandle
	c := newConn(h, hh)
	h.connsmu.Lock()
	h.conns[hh] = c
	h.connsmu.Unlock()
	h.setAdvertiseEnable(true)

	// FIXME: sloppiness. This call should be called by the package user once we
	// flesh out the support of l2cap signaling packets (CID:0x0001,0x0005)
	if ep.ConnLatency != 0 || ep.ConnInterval > 0x18 {
		c.updateConnection()
	}

	// master connection
	if ep.Role == 0x01 {
		pd := &PlatData{
			Address: ep.PeerAddress,
			Conn:    c,
		}
		h.AcceptMasterHandler(pd)
		return
	}
	h.plistmu.Lock()
	if pd := h.plist[ep.PeerAddress]; pd != nil {
		h.plistmu.Unlock()
		pd.Conn = c
		h.AcceptSlaveHandler(pd)
	} else {
		log.Printf("HCI: can't find data for %v", ep.PeerAddress)
	}
}

func (h *HCI) handleDisconnectionComplete(b []byte) error {
	ep := &evt.DisconnectionCompleteEP{}
	if err := ep.Unmarshal(b); err != nil {
		return err
	}
	hh := ep.ConnectionHandle
	h.connsmu.Lock()
	defer h.connsmu.Unlock()
	c, found := h.conns[hh]
	if !found {
		// should not happen, just be cautious for now.
		log.Printf("l2conn: disconnecting a disconnected 0x%04X connection", hh)
		return nil
	}
	delete(h.conns, hh)
	close(c.aclc)
	h.setAdvertiseEnable(true)
	return nil
}

func (h *HCI) handleLTKRequest(b []byte) {
	ep := &evt.LELTKRequestEP{}
	if err := ep.Unmarshal(b); err != nil {
		log.Printf("ltkrequest: error, parsing request")
		return
	}
	hh := ep.ConnectionHandle
	h.connsmu.Lock()
	defer h.connsmu.Unlock()
	_, found := h.conns[hh]
	if !found {
		// should not happen, just be cautious for now.
		log.Printf("ltkrequest: error, connection 0x%04X probably expired", hh)
		return
	}
	h.c.Send(cmd.LELTKNegReply{ConnectionHandle: hh})
	// TODO: implement proper key management
}

func (h *HCI) handleLEMeta(b []byte) error {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("error while handling meta: %v", err)
		}
	}()

	code := evt.LEEventCode(b[0])
	switch code {
	case evt.LEConnectionComplete:
		go h.handleConnection(b)
	case evt.LEConnectionUpdateComplete:
		// anything to do here?
	case evt.LEAdvertisingReport:
		go h.handleAdvertisement(b)
	// case evt.LEReadRemoteUsedFeaturesComplete:
	case evt.LELTKRequest:
		go h.handleLTKRequest(b)
	// case evt.LERemoteConnectionParameterRequest:
	default:
		return fmt.Errorf("Unhandled LE event: 0x%02x, [ % X ]", code, b)
	}
	return nil
}

func (h *HCI) handleL2CAP(b []byte) error {
	a := &aclData{}
	if err := a.unmarshal(b); err != nil {
		return err
	}
	h.connsmu.Lock()
	defer h.connsmu.Unlock()
	c, found := h.conns[a.attr]
	if !found {
		// should not happen, just be cautious for now.
		log.Printf("l2conn: got data for disconnected handle: 0x%04x", a.attr)
		return nil
	}
	c.aclc <- a
	return nil
}
