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

var flagNames = []string{
	"UP",
	"BROADCAST",
	"LOOPBACK",
	"POINT2POINT",
	"MULTICAST",
}

type addrJSON struct {
	Address string `json:"address"`
	Type    string `json:"type"`
}

type ifaceJSON struct {
	Index     int        `json:"index"`
	MTU       int        `json:"mtu"`
	Name      string     `json:"name"`
	MAC       string     `json:"mac"`
	Vendor    string     `json:"vendor"`
	Flags     []string   `json:"flags"`
	Addresses []addrJSON `json:"addresses"`
}

type resourcesJSON struct {
	NumCPU       int    `json:"cpus"`
	MaxCPU       int    `json:"max_cpus"`
	NumGoroutine int    `json:"goroutines"`
	Alloc        uint64 `json:"alloc"`
	Sys          uint64 `json:"sys"`
	NumGC        uint32 `json:"gcs"`
}

type SessionJSON struct {
	Version    string            `json:"version"`
	OS         string            `json:"os"`
	Arch       string            `json:"arch"`
	GoVersion  string            `json:"goversion"`
	Resources  resourcesJSON     `json:"resources"`
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
	PolledAt   time.Time         `json:"polled_at"`
	Active     bool              `json:"active"`
	GPS        GPS               `json:"gps"`
	Modules    ModuleList        `json:"modules"`
	Caplets    []*caplets.Caplet `json:"caplets"`
}

func (s *Session) MarshalJSON() ([]byte, error) {
	var m runtime.MemStats

	runtime.ReadMemStats(&m)

	doc := SessionJSON{
		Version:   core.Version,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
		Resources: resourcesJSON{
			NumCPU:       runtime.NumCPU(),
			MaxCPU:       runtime.GOMAXPROCS(0),
			NumGoroutine: runtime.NumGoroutine(),
			Alloc:        m.Alloc,
			Sys:          m.Sys,
			NumGC:        m.NumGC,
		},
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
		PolledAt:   time.Now(),
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
		mac := network.NormalizeMac(iface.HardwareAddr.String())

		ij := ifaceJSON{
			Index:     iface.Index,
			MTU:       iface.MTU,
			Name:      iface.Name,
			MAC:       mac,
			Vendor:    network.ManufLookup(mac),
			Flags:     make([]string, 0),
			Addresses: make([]addrJSON, 0),
		}

		if addrs, err := iface.Addrs(); err == nil {
			for _, addr := range addrs {
				ij.Addresses = append(ij.Addresses, addrJSON{
					Address: addr.String(),
					Type:    addr.Network(),
				})
			}
		}

		for bit, name := range flagNames {
			if iface.Flags&(1<<uint(bit)) != 0 {
				ij.Flags = append(ij.Flags, name)
			}
		}

		doc.Interfaces = append(doc.Interfaces, ij)
	}

	return json.Marshal(doc)
}
