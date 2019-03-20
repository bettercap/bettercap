package session

import (
	"encoding/json"
	"net"
	"runtime"
	"time"

	"github.com/bettercap/bettercap/caplets"
	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"
)

type ifaceJSON struct {
	Index int       `json:"index"`
	MTU   int       `json:"mtu"`
	Name  string    `json:"name"`
	MAC   string    `json:"mac"`
	Flags net.Flags `json:"flags"`
}

type sessionJSON struct {
	Version    string            `json:"version"`
	OS         string            `json:"os"`
	Arch       string            `json:"arch"`
	GoVersion  string            `json:"goversion"`
	Interfaces []ifaceJSON       `json:"interfaces"`
	Options    core.Options      `json:"options"`
	Interface  *network.Endpoint `json:"interface"`
	Gateway    *network.Endpoint `json:"gateway"`
	Env        *Environment      `json:"env"`
	Lan        *network.LAN      `json:"lan"`
	WiFi       *network.WiFi     `json:"wifi"`
	BLE        *network.BLE      `json:"ble"`
	HID        *network.HID      `json:"hid"`
	Queue      *packets.Queue    `json:"packets"`
	StartedAt  time.Time         `json:"started_at"`
	Active     bool              `json:"active"`
	GPS        GPS               `json:"gps"`
	Modules    ModuleList        `json:"modules"`
	Caplets    []*caplets.Caplet `json:"caplets"`
}

func (s *Session) MarshalJSON() ([]byte, error) {
	doc := sessionJSON{
		Version:    core.Version,
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		GoVersion:  runtime.Version(),
		Interfaces: make([]ifaceJSON, 0),
		Options:    s.Options,
		Interface:  s.Interface,
		Gateway:    s.Gateway,
		Env:        s.Env,
		Lan:        s.Lan,
		WiFi:       s.WiFi,
		BLE:        s.BLE,
		HID:        s.HID,
		Queue:      s.Queue,
		StartedAt:  s.StartedAt,
		Active:     s.Active,
		GPS:        s.GPS,
		Modules:    s.Modules,
		Caplets:    caplets.List(),
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		doc.Interfaces = append(doc.Interfaces, ifaceJSON{
			Index: iface.Index,
			MTU:   iface.MTU,
			Name:  iface.Name,
			MAC:   iface.HardwareAddr.String(),
			Flags: iface.Flags,
		})
	}

	return json.Marshal(doc)
}
