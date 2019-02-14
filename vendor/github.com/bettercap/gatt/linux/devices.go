package linux

import "github.com/bettercap/gatt/linux/gioctl"

const (
	ioctlSize     = uintptr(4)
	hciMaxDevices = 16
	typHCI        = 72 // 'H'
)

var (
	hciUpDevice      = gioctl.IoW(typHCI, 201, ioctlSize) // HCIDEVUP
	hciDownDevice    = gioctl.IoW(typHCI, 202, ioctlSize) // HCIDEVDOWN
	hciResetDevice   = gioctl.IoW(typHCI, 203, ioctlSize) // HCIDEVRESET
	hciGetDeviceList = gioctl.IoR(typHCI, 210, ioctlSize) // HCIGETDEVLIST
	hciGetDeviceInfo = gioctl.IoR(typHCI, 211, ioctlSize) // HCIGETDEVINFO
)

type devRequest struct {
	id  uint16
	opt uint32
}

type devListRequest struct {
	devNum     uint16
	devRequest [hciMaxDevices]devRequest
}

type hciDevInfo struct {
	id         uint16
	name       [8]byte
	bdaddr     [6]byte
	flags      uint32
	devType    uint8
	features   [8]uint8
	pktType    uint32
	linkPolicy uint32
	linkMode   uint32
	aclMtu     uint16
	aclPkts    uint16
	scoMtu     uint16
	scoPkts    uint16

	stats hciDevStats
}

type hciDevStats struct {
	errRx  uint32
	errTx  uint32
	cmdTx  uint32
	evtRx  uint32
	aclTx  uint32
	aclRx  uint32
	scoTx  uint32
	scoRx  uint32
	byteRx uint32
	byteTx uint32
}
