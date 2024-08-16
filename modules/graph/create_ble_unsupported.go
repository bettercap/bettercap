//go:build windows
// +build windows

package graph

import "github.com/bettercap/bettercap/v2/network"

func (mod *Module) createBLEServerGraph(dev *network.BLEDevice) (*Node, bool, error) {
	return nil, false, nil
}
