package session

import (
	"fmt"
	"sort"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/net"
)

type tSorter []*net.Endpoint

func (a tSorter) Len() int           { return len(a) }
func (a tSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a tSorter) Less(i, j int) bool { return a[i].IpAddressUint32 < a[j].IpAddressUint32 }

func (tp *Targets) Dump() {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	fmt.Println()
	fmt.Printf("  " + core.GREEN + "interface" + core.RESET + "\n\n")
	fmt.Printf("    " + tp.Interface.String() + "\n")
	fmt.Println()
	fmt.Printf("  " + core.GREEN + "gateway" + core.RESET + "\n\n")
	fmt.Printf("    " + tp.Gateway.String() + "\n")

	if len(tp.Targets) > 0 {
		fmt.Println()
		fmt.Printf("  " + core.GREEN + "hosts" + core.RESET + "\n\n")
		targets := make([]*net.Endpoint, 0, len(tp.Targets))
		for _, t := range tp.Targets {
			targets = append(targets, t)
		}

		sort.Sort(tSorter(targets))

		for _, t := range targets {
			fmt.Println("    " + t.String())
		}
	}

	fmt.Println()
}
