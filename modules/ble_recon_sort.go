// +build !windows
// +build !darwin

package modules

import (
	"github.com/bettercap/bettercap/network"
)

type ByBLERSSISorter []*network.BLEDevice

func (a ByBLERSSISorter) Len() int      { return len(a) }
func (a ByBLERSSISorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByBLERSSISorter) Less(i, j int) bool {
	return a[i].RSSI > a[j].RSSI
}
