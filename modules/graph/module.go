package graph

import (
	"fmt"
	"github.com/bettercap/bettercap/caplets"
	"github.com/bettercap/bettercap/modules/wifi"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/session"
	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/str"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const (
	ifaceAnnotation = "<interface>"
	edgeStaleTime   = time.Hour * 24
)

type settings struct {
	path string
	layout string
	name string
	output string
}

type Module struct {
	session.SessionModule

	settings settings
	db       *Graph
	gw       *Node
	iface    *Node
	eventBus session.EventBus
}

func NewModule(s *session.Session) *Module {
	mod := &Module{
		SessionModule: session.NewSessionModule("graph", s),
		settings: settings{
			path: filepath.Join(caplets.InstallBase, "graph"),
			layout: "neato",
			name: "bettergraph",
			output: "bettergraph.dot",
		},
	}

	mod.AddParam(session.NewStringParameter("graph.path",
		mod.settings.path,
		"",
		"Base path for the graph database."))

	mod.AddParam(session.NewStringParameter("graph.dot.name",
		mod.settings.name,
		"",
		"Graph name in the dot output."))

	mod.AddParam(session.NewStringParameter("graph.dot.layout",
		mod.settings.layout,
		"",
		"Layout for dot output."))

	mod.AddParam(session.NewStringParameter("graph.dot.output",
		mod.settings.output,
		"",
		"File name for dot output."))

	mod.AddHandler(session.NewModuleHandler("graph on", "",
		"Start the Module module.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("graph off", "",
		"Stop the Module module.",
		func(args []string) error {
			return mod.Stop()
		}))

	mod.AddHandler(session.NewModuleHandler("graph.to_dot MAC?",
		`graph\.to_dot\s*([^\s]*)`,
		"Generate a dot graph file from the current graph.",
		func(args []string) (err error) {
			bssid := ""
			if len(args) == 1 && args[0] != "" {
				bssid = network.NormalizeMac(str.Trim(args[0]))
			}
			return mod.generateDotGraph(bssid)
		}))

	return mod
}

func (mod *Module) Name() string {
	return "graph"
}

func (mod *Module) Description() string {
	return "A module to build a graph of WiFi and LAN nodes."
}

func (mod *Module) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *Module) updateSettings() error {
	var err error

	if err, mod.settings.name = mod.StringParam("graph.dot.name"); err != nil {
		return err
	} else if err, mod.settings.layout = mod.StringParam("graph.dot.layout"); err != nil {
		return err
	} else if err, mod.settings.output = mod.StringParam("graph.dot.output"); err != nil {
		return err
	} else if err, mod.settings.path = mod.StringParam("graph.path"); err != nil {
		return err
	} else if mod.settings.path, err = filepath.Abs(mod.settings.path); err != nil {
		return err
	} else if !fs.Exists(mod.settings.path) {
		if err = os.MkdirAll(mod.settings.path, os.ModePerm); err != nil {
			return err
		}
	}

	if mod.db, err = NewGraph(mod.settings.path); err != nil {
		return err
	}

	return nil
}

func (mod *Module) Configure() (err error) {
	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	} else if err = mod.updateSettings(); err != nil {
		return err
	}

	// if have an IP
	if mod.Session.Gateway != nil && mod.Session.Interface != nil {
		// find or create gateway node
		gw := mod.Session.Gateway
		if mod.gw, err = mod.db.FindNode(Gateway, gw.HwAddress); err != nil {
			return err
		} else if mod.gw == nil {
			if mod.gw, err = mod.db.CreateNode(Gateway, gw.HwAddress, gw, ""); err != nil {
				return err
			}
		} else {
			if err = mod.db.UpdateNode(mod.gw); err != nil {
				return err
			}
		}

		// find or create interface node
		iface := mod.Session.Interface
		if mod.iface, err = mod.db.FindNode(Endpoint, iface.HwAddress); err != nil {
			return err
		} else if mod.iface == nil {
			// create the interface node
			if mod.iface, err = mod.db.CreateNode(Endpoint, iface.HwAddress, iface, ifaceAnnotation); err != nil {
				return err
			}
		} else if err = mod.db.UpdateNode(mod.iface); err != nil {
			return err
		}

		// create relations if needed
		if manages, err := mod.db.FindLastRecentEdgeOfType(mod.gw, mod.iface, Manages, edgeStaleTime); err != nil {
			return err
		} else if manages == nil {
			if manages, err = mod.db.CreateEdge(mod.gw, mod.iface, Manages); err != nil {
				return err
			}
		}

		if connects_to, err := mod.db.FindLastEdgeOfType(mod.iface, mod.gw, ConnectsTo); err != nil {
			return err
		} else if connects_to == nil {
			if connects_to, err = mod.db.CreateEdge(mod.iface, mod.gw, ConnectsTo); err != nil {
				return err
			}
		}
	}

	mod.eventBus = mod.Session.Events.Listen()

	return nil
}

func (mod *Module) generateDotGraph(bssid string) error {
	start := time.Now()

	if err := mod.updateSettings(); err != nil {
		return err
	} else if data, err := mod.db.Dot(bssid, mod.settings.layout, mod.settings.name); err != nil {
		return err
	} else if err := ioutil.WriteFile(mod.settings.output, []byte(data), os.ModePerm); err != nil {
		return err
	}

	mod.Info("graph saved to %s in %v", mod.settings.output, time.Since(start))
	return nil
}

func (mod *Module) createIPGraph(endpoint *network.Endpoint) (*Node, bool, error) {
	node, err := mod.db.FindNode(Endpoint, endpoint.HwAddress)
	isNew := node == nil
	if err != nil {
		return nil, false, err
	} else if isNew {
		if node, err = mod.db.CreateNode(Endpoint, endpoint.HwAddress, endpoint, ""); err != nil {
			return nil, false, err
		}
	} else {
		if err = mod.db.UpdateNode(node); err != nil {
			return nil, false, err
		}
	}

	// create relations if needed
	if manages, err := mod.db.FindLastRecentEdgeOfType(mod.gw, node, Manages, edgeStaleTime); err != nil {
		return nil, false, err
	} else if manages == nil {
		if manages, err = mod.db.CreateEdge(mod.gw, node, Manages); err != nil {
			return nil, false, err
		}
	}

	if connects_to, err := mod.db.FindLastRecentEdgeOfType(node, mod.gw, ConnectsTo, edgeStaleTime); err != nil {
		return nil, false, err
	} else if connects_to == nil {
		if connects_to, err = mod.db.CreateEdge(node, mod.gw, ConnectsTo); err != nil {
			return nil, false, err
		}
	}

	return node, isNew, nil
}

func (mod *Module) createDot11ApGraph(ap *network.AccessPoint) (*Node, bool, error) {
	node, err := mod.db.FindNode(AccessPoint, ap.HwAddress)
	isNew := node == nil
	if err != nil {
		return nil, false, err
	} else if isNew {
		if node, err = mod.db.CreateNode(AccessPoint, ap.HwAddress, ap, ""); err != nil {
			return nil, false, err
		}
	} else if err = mod.db.UpdateNode(node); err != nil {
		return nil, false, err
	}
	return node, isNew, nil
}

func (mod *Module) createDot11SSIDGraph(hw string, ssid string) (*Node, bool, error) {
	node, err := mod.db.FindNode(SSID, hw)
	isNew := node == nil
	if err != nil {
		return nil, false, err
	} else if isNew {
		if node, err = mod.db.CreateNode(SSID, hw, ssid, ""); err != nil {
			return nil, false, err
		}
	} else if err = mod.db.UpdateNode(node); err != nil {
		return nil, false, err
	}
	return node, isNew, nil
}

func (mod *Module) createDot11StaGraph(station *network.Station) (*Node, bool, error) {
	node, err := mod.db.FindNode(Station, station.HwAddress)
	isNew := node == nil
	if err != nil {
		return nil, false, err
	} else if isNew {
		if node, err = mod.db.CreateNode(Station, station.HwAddress, station, ""); err != nil {
			return nil, false, err
		}
	} else if err = mod.db.UpdateNode(node); err != nil {
		return nil, false, err
	}
	return node, isNew, nil
}

func (mod *Module) createDot11Graph(ap *network.AccessPoint, station *network.Station) (*Node, bool, *Node, bool, error) {
	apNode, apIsNew, err := mod.createDot11ApGraph(ap)
	if err != nil {
		return nil, false, nil, false, err
	}

	staNode, staIsNew, err := mod.createDot11StaGraph(station)
	if err != nil {
		return nil, false, nil, false, err
	}

	// create relations if needed
	if manages, err := mod.db.FindLastRecentEdgeOfType(apNode, staNode, Manages, edgeStaleTime); err != nil {
		return nil, false, nil, false, err
	} else if manages == nil {
		if manages, err = mod.db.CreateEdge(apNode, staNode, Manages); err != nil {
			return nil, false, nil, false, err
		}
	}

	if connects_to, err := mod.db.FindLastRecentEdgeOfType(staNode, apNode, ConnectsTo, edgeStaleTime); err != nil {
		return nil, false, nil, false, err
	} else if connects_to == nil {
		if connects_to, err = mod.db.CreateEdge(staNode, apNode, ConnectsTo); err != nil {
			return nil, false, nil, false, err
		}
	}

	return apNode, apIsNew, staNode, staIsNew, nil
}

func (mod *Module) createDot11ProbeGraph(ssid string, station *network.Station) (*Node, bool, *Node, bool, error) {
	apNode, apIsNew, err := mod.createDot11SSIDGraph(station.HwAddress + fmt.Sprintf(":PROBE:%x", ssid), ssid)
	if err != nil {
		return nil, false, nil, false, err
	}

	staNode, staIsNew, err := mod.createDot11StaGraph(station)
	if err != nil {
		return nil, false, nil, false, err
	}

	// create relations if needed
	if probes_for, err := mod.db.FindLastRecentEdgeOfType(staNode, apNode, ProbesFor, edgeStaleTime); err != nil {
		return nil, false, nil, false, err
	} else if probes_for == nil {
		if probes_for, err = mod.db.CreateEdge(staNode, apNode, ProbesFor); err != nil {
			return nil, false, nil, false, err
		}
	}

	return apNode, apIsNew, staNode, staIsNew, nil
}

func (mod *Module) createBLEServerGraph(dev *network.BLEDevice) (*Node, bool, error) {
	mac := network.NormalizeMac(dev.Device.ID())
	node, err := mod.db.FindNode(BLEServer, mac)
	isNew := node == nil
	if err != nil {
		return nil, false, err
	} else if isNew {
		if node, err = mod.db.CreateNode(BLEServer, mac, dev, ""); err != nil {
			return nil, false, err
		}
	} else if err = mod.db.UpdateNode(node); err != nil {
		return nil, false, err
	}
	return node, isNew, nil
}

func (mod *Module) connectAsSame(a, b *Node) error {
	if aIsB, err := mod.db.FindLastEdgeOfType(a, b, Is); err != nil {
		return err
	} else if aIsB == nil {
		if aIsB, err = mod.db.CreateEdge(a, b, Is); err != nil {
			return err
		}
	}

	if bIsA, err := mod.db.FindLastEdgeOfType(b, a, Is); err != nil {
		return err
	} else if bIsA == nil {
		if bIsA, err = mod.db.CreateEdge(b, a, Is); err != nil {
			return err
		}
	}

	return nil
}

func (mod *Module) onEvent(e session.Event) {
	var entities []*Node

	if e.Tag == "endpoint.new" {
		endpoint := e.Data.(*network.Endpoint)
		if entity, _, err := mod.createIPGraph(endpoint); err != nil {
			mod.Error("%s", err)
		} else {
			entities = append(entities, entity)
		}
	} else if e.Tag == "wifi.ap.new" {
		ap := e.Data.(*network.AccessPoint)
		if entity, _, err := mod.createDot11ApGraph(ap); err != nil {
			mod.Error("%s", err)
		} else {
			entities = append(entities, entity)
		}
	} else if e.Tag == "wifi.client.new" {
		ce := e.Data.(wifi.ClientEvent)
		if apEntity, _, staEntity, _, err := mod.createDot11Graph(ce.AP, ce.Client); err != nil {
			mod.Error("%s", err)
		} else {
			entities = append(entities, apEntity)
			entities = append(entities, staEntity)
		}
	} else if e.Tag == "wifi.client.probe" {
		probe := e.Data.(wifi.ProbeEvent)
		station := network.Station{
			RSSI: probe.RSSI,
			Endpoint: &network.Endpoint{
				HwAddress: probe.FromAddr,
				Vendor:    probe.FromVendor,
				Alias:     probe.FromAlias,
			},
		}

		if _, _, staEntity, _, err := mod.createDot11ProbeGraph(probe.SSID, &station); err != nil {
			mod.Error("%s", err)
		} else {
			// don't add fake ap to entities, no need to correlate
			entities = append(entities, staEntity)
		}
	} else if e.Tag == "ble.device.new" {
		// surprisingly some devices, like DLink IPCams have BLE, Dot11 and LAN hardware address in common
		dev := e.Data.(*network.BLEDevice)
		if entity, _, err := mod.createBLEServerGraph(dev); err != nil {
			mod.Error("%s", err)
		} else {
			entities = append(entities, entity)
		}
	}

	// if there's at least an entity node, search for other nodes with the
	// same mac address but different type and connect them as needed
	for _, entity := range entities {
		if others, err := mod.db.FindOtherTypes(entity.Type, entity.ID); err != nil {
			mod.Error("%s", err)
		} else if len(others) > 0 {
			for _, other := range others {
				if err = mod.connectAsSame(entity, other); err != nil {
					mod.Error("%s", err)
				}
			}
		}
	}
}

func (mod *Module) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.Info("started with database @ %s", mod.settings.path)

		for mod.Running() {
			select {
			case e := <-mod.eventBus:
				mod.onEvent(e)
			}
		}
	})
}

func (mod *Module) Stop() error {
	return mod.SetRunning(false, func() {
		mod.Session.Events.Unlisten(mod.eventBus)
	})
}
