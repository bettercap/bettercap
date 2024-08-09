package graph

import (
	"fmt"

	"github.com/bettercap/bettercap/v2/network"
)

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

func (mod *Module) createDot11SSIDGraph(hex string, ssid string) (*Node, bool, error) {
	node, err := mod.db.FindNode(SSID, hex)
	isNew := node == nil
	if err != nil {
		return nil, false, err
	} else if isNew {
		if node, err = mod.db.CreateNode(SSID, hex, ssid, ""); err != nil {
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
	ssidNode, ssidIsNew, err := mod.createDot11SSIDGraph(fmt.Sprintf("%x", ssid), ssid)
	if err != nil {
		return nil, false, nil, false, err
	}

	staNode, staIsNew, err := mod.createDot11StaGraph(station)
	if err != nil {
		return nil, false, nil, false, err
	}

	// create relations if needed
	if probes_for, err := mod.db.FindLastRecentEdgeOfType(staNode, ssidNode, ProbesFor, edgeStaleTime); err != nil {
		return nil, false, nil, false, err
	} else if probes_for == nil {
		if probes_for, err = mod.db.CreateEdge(staNode, ssidNode, ProbesFor); err != nil {
			return nil, false, nil, false, err
		}
	}

	if probed_by, err := mod.db.FindLastRecentEdgeOfType(ssidNode, staNode, ProbedBy, edgeStaleTime); err != nil {
		return nil, false, nil, false, err
	} else if probed_by == nil {
		if probed_by, err = mod.db.CreateEdge(ssidNode, staNode, ProbedBy); err != nil {
			return nil, false, nil, false, err
		}
	}

	return ssidNode, ssidIsNew, staNode, staIsNew, nil
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
