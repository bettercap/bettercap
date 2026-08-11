package wifi

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	"github.com/bettercap/bettercap/v2/network"
	"github.com/gobwas/glob"
)

type wifiSelector struct {
	raw     string
	all     bool
	address net.HardwareAddr
	essid   glob.Glob
}

type wifiTarget struct {
	ap             *network.AccessPoint
	client         *network.Station
	apSnapshot     network.StationSnapshot
	clientSnapshot network.StationSnapshot
}

func hardwareAddrIn(addresses []net.HardwareAddr, target net.HardwareAddr) bool {
	for _, address := range addresses {
		if bytes.Equal(address, target) {
			return true
		}
	}
	return false
}

func (mod *WiFiModule) parseWiFiSkipList(name string) ([]net.HardwareAddr, error) {
	err, value := mod.StringParam(name)
	if err != nil {
		return nil, err
	}
	return network.ParseMACs(value)
}

func newWiFiSelector(target string) (wifiSelector, error) {
	selector := wifiSelector{raw: target}
	// Preserve the legacy selector precedence: all aliases and valid MAC
	// addresses must never be interpreted as ESSIDs or glob expressions.
	if target == "all" || target == "*" {
		selector.all = true
		return selector, nil
	}

	if address, err := net.ParseMAC(target); err == nil {
		selector.address = address
		selector.all = network.IsBroadcastMac(address)
		return selector, nil
	}

	matcher, err := glob.Compile(target)
	if err != nil {
		return wifiSelector{}, fmt.Errorf("invalid ESSID expression %q: %w", target, err)
	}
	selector.essid = matcher
	return selector, nil
}

func (selector wifiSelector) matchesAP(snapshot network.StationSnapshot) bool {
	return selector.all ||
		(selector.address != nil && bytes.Equal(snapshot.HW, selector.address)) ||
		(selector.essid != nil && selector.essid.Match(snapshot.Hostname))
}

func (selector wifiSelector) matchesClient(snapshot network.StationSnapshot) bool {
	return selector.all || (selector.address != nil && bytes.Equal(snapshot.HW, selector.address))
}

func (mod *WiFiModule) resolveWiFiTargets(selector wifiSelector, includeClients bool) []wifiTarget {
	targets := make([]wifiTarget, 0)
	for _, ap := range mod.Session.WiFi.List() {
		apSnapshot := ap.Snapshot()
		matchesAP := selector.matchesAP(apSnapshot)
		if !includeClients {
			if matchesAP {
				targets = append(targets, wifiTarget{ap: ap, apSnapshot: apSnapshot})
			}
			continue
		}

		for _, client := range ap.Clients() {
			clientSnapshot := client.Snapshot()
			if matchesAP || selector.matchesClient(clientSnapshot) {
				targets = append(targets, wifiTarget{
					ap:             ap,
					client:         client,
					apSnapshot:     apSnapshot,
					clientSnapshot: clientSnapshot,
				})
			}
		}
	}
	return targets
}

func (mod *WiFiModule) wifiTargetCompleter(prefix string, includeClients bool) []string {
	results := []string{""}
	seen := make(map[string]bool)
	add := func(value string) {
		if !seen[value] && (prefix == "" || strings.HasPrefix(value, prefix)) {
			seen[value] = true
			results = append(results, value)
		}
	}

	add("all")
	add("*")
	for _, ap := range mod.Session.WiFi.List() {
		snapshot := ap.Snapshot()
		add(snapshot.HwAddress)
		if snapshot.Hostname != "" {
			add(snapshot.Hostname)
		}
		if includeClients {
			for _, client := range ap.Clients() {
				add(client.BSSID())
			}
		}
	}
	return results
}
