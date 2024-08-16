//go:build !windows
// +build !windows

package graph

import "github.com/bettercap/bettercap/v2/network"

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
