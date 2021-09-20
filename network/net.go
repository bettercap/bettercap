package network

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/bettercap/bettercap/core"

	"github.com/evilsocket/islazy/data"
	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/islazy/tui"

	"github.com/malfunkt/iprange"
)

var ErrNoIfaces = errors.New("No active interfaces found.")
var ErrNoGateway = errors.New("Could not detect gateway.")

const (
	MonitorModeAddress = "0.0.0.0"
	BroadcastSuffix    = ".255"
	BroadcastMac       = "ff:ff:ff:ff:ff:ff"
	IPv4MulticastStart = "01:00:5e:00:00:00"
	IPv4MulticastEnd   = "01:00:5e:7f:ff:ff"
)

var (
	BroadcastHw        = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	IPv4BlockValidator = regexp.MustCompile(`^` +
		`(?:(?:25[0-5]|2[0-4][0-9]|[1][0-9]{2}|[1-9]?[0-9])\.){3}` +
		`(?:25[0-5]|2[0-4][0-9]|[1][0-9]{2}|[1-9]?[0-9])` +
		`/(?:3[0-2]|2[0-9]|[1]?[0-9])` + `$`)
	IPv4RangeValidator = regexp.MustCompile(`^` +
		`(?:(?:(?:25[0-5]|2[0-4][0-9]|[1][0-9]{2}|[1-9]?[0-9])-)?(?:25[0-5]|2[0-4][0-9]|[1][0-9]{2}|[1-9]?[0-9])\.){3}` +
		`(?:(?:25[0-5]|2[0-4][0-9]|[1][0-9]{2}|[1-9]?[0-9])-)?(?:25[0-5]|2[0-4][0-9]|[1][0-9]{2}|[1-9]?[0-9])` + `$`)
	IPv4Validator = regexp.MustCompile(`^` +
		`(?:(?:25[0-5]|2[0-4][0-9]|[1][0-9]{2}|[1-9]?[0-9])\.){3}` +
		`(?:25[0-5]|2[0-4][0-9]|[1][0-9]{2}|[1-9]?[0-9])` + `$`)
	MACValidator = regexp.MustCompile(`(?i)^(?:[a-f0-9]{2}:){5}[a-f0-9]{2}$`)
	// lulz this sounds like a hamburger
	macParser   = regexp.MustCompile(`(?i)((?:[a-f0-9]{2}:){5}[a-f0-9]{2})`)
	aliasParser = regexp.MustCompile(`(?i)([a-z_][a-z_0-9]+)`)
)

func IsZeroMac(mac net.HardwareAddr) bool {
	for _, b := range mac {
		if b != 0x00 {
			return false
		}
	}
	return true
}

func IsBroadcastMac(mac net.HardwareAddr) bool {
	for _, b := range mac {
		if b != 0xff {
			return false
		}
	}
	return true
}

func NormalizeMac(mac string) string {
	var parts []string
	if strings.ContainsRune(mac, '-') {
		parts = strings.Split(mac, "-")
	} else {
		parts = strings.Split(mac, ":")
	}

	for i, p := range parts {
		if len(p) < 2 {
			parts[i] = "0" + p
		}
	}
	return strings.ToLower(strings.Join(parts, ":"))
}

func ParseMACs(targets string) (macs []net.HardwareAddr, err error) {
	macs = make([]net.HardwareAddr, 0)
	if targets = str.Trim(targets); targets == "" {
		return
	}

	for _, mac := range macParser.FindAllString(targets, -1) {
		mac = NormalizeMac(mac)
		hw, err := net.ParseMAC(mac)
		if err != nil {
			return nil, fmt.Errorf("error while parsing MAC '%s': %s", mac, err)
		}

		macs = append(macs, hw)
		targets = strings.Replace(targets, mac, "", -1)
	}

	return
}

func ParseTargets(targets string, aliasMap *data.UnsortedKV) (ips []net.IP, macs []net.HardwareAddr, err error) {
	ips = make([]net.IP, 0)
	macs = make([]net.HardwareAddr, 0)

	if targets = str.Trim(targets); targets == "" {
		return
	}

	// first isolate MACs and parse them
	for _, mac := range macParser.FindAllString(targets, -1) {
		normalizedMac := NormalizeMac(mac)
		hw, err := net.ParseMAC(normalizedMac)
		if err != nil {
			return nil, nil, fmt.Errorf("error while parsing MAC '%s': %s", normalizedMac, err)
		}

		macs = append(macs, hw)
		targets = strings.Replace(targets, mac, "", -1)
	}
	targets = strings.Trim(targets, ", ")

	// fmt.Printf("targets=%s macs=%#v\n", targets, macs)

	// check and resolve aliases
	for _, targetAlias := range aliasParser.FindAllString(targets, -1) {
		found := false
		aliasMap.Each(func(mac, alias string) bool {
			if alias == targetAlias {
				if hw, err := net.ParseMAC(mac); err == nil {
					macs = append(macs, hw)
					targets = strings.Replace(targets, targetAlias, "", -1)
					found = true
					return true
				}
			}
			return false
		})

		if !found {
			return nil, nil, fmt.Errorf("could not resolve alias %s", targetAlias)
		}
	}
	targets = strings.Trim(targets, ", ")

	// parse what's left
	if targets != "" {
		list, err := iprange.ParseList(targets)
		if err != nil {
			return nil, nil, fmt.Errorf("error while parsing address list '%s': %s", targets, err)
		}

		ips = list.Expand()
	}

	return
}

func ParseEndpoints(targets string, lan *LAN) ([]*Endpoint, error) {
	ips, macs, err := ParseTargets(targets, lan.Aliases())
	if err != nil {
		return nil, err
	}

	tmp := make(map[string]*Endpoint)
	for _, ip := range ips {
		if e := lan.GetByIp(ip.String()); e != nil {
			tmp[e.HW.String()] = e
		}
	}

	for _, mac := range macs {
		if e, found := lan.Get(mac.String()); found {
			tmp[e.HW.String()] = e
		}
	}

	ret := make([]*Endpoint, 0)
	for _, e := range tmp {
		ret = append(ret, e)
	}
	return ret, nil
}

func buildEndpointFromInterface(iface net.Interface) (*Endpoint, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	ifName := getInterfaceName(iface)

	e := NewEndpointNoResolve(MonitorModeAddress, iface.HardwareAddr.String(), ifName, 0)

	e.Index = iface.Index

	for _, a := range addrs {
		address := a.String()
		switch true {
		case IPv4Validator.MatchString(address):
			e.SetIP(address)
			break
		case IPv4BlockValidator.MatchString(address):
			e.SetNetwork(address)
			break
		default:
			e.SetIPv6(address)
			break
		}
	}

	return e, nil
}

func matchByAddress(iface net.Interface, name string) bool {
	ifMac := iface.HardwareAddr.String()
	if NormalizeMac(ifMac) == NormalizeMac(name) {
		return true
	}

	addrs, err := iface.Addrs()
	if err == nil {
		for _, addr := range addrs {
			ip := addr.String()
			if ip == name || strings.HasPrefix(ip, name) {
				return true
			}
		}
	}

	return false
}

func findInterfaceByName(name string, ifaces []net.Interface) (*Endpoint, error) {
	for _, iface := range ifaces {
		ifName := getInterfaceName(iface)
		if ifName == name || matchByAddress(iface, name) {
			return buildEndpointFromInterface(iface)
		}
	}

	return nil, fmt.Errorf("no interface matching '%s' found.", name)
}

func FindInterface(name string) (*Endpoint, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	name = str.Trim(name)
	if name != "" {
		return findInterfaceByName(name, ifaces)
	}

	// user did not provide an interface name,
	// return the first one with a valid ipv4
	// address that does not loop back
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Printf("wtf of the day: %s", err)
			continue
		}

		for _, address := range addrs {
			ip := address.String()
			if !strings.HasPrefix(ip, "127.0.0.1") && IPv4BlockValidator.MatchString(ip) {
				return buildEndpointFromInterface(iface)
			}
		}
	}

	return nil, ErrNoIfaces
}

func SetWiFiRegion(region string) error {
	if core.HasBinary("iw") {
		if out, err := core.Exec("iw", []string{"reg", "set", region}); err != nil {
			return err
		} else if out != "" {
			return fmt.Errorf("unexpected output while setting WiFi region %s: %s", region, out)
		}
	}
	return nil
}

func ActivateInterface(name string) error {
	if out, err := core.Exec("ifconfig", []string{name, "up"}); err != nil {
		if out != "" {
			return fmt.Errorf("%v: %s", err, out)
		} else {
			return err
		}
	} else if out != "" {
		return fmt.Errorf("unexpected output while activating interface %s: %s", name, out)
	}
	return nil
}

func SetInterfaceTxPower(name string, txpower int) error {
	if core.HasBinary("iw") {
		Debug("SetInterfaceTxPower(%s, %d) iw based", name, txpower)
		if _, err := core.Exec("iw", []string{"dev", name, "set", "txpower", "fixed", fmt.Sprintf("%d", txpower)}); err != nil {
			return err
		}
	} else if core.HasBinary("iwconfig") {
		Debug("SetInterfaceTxPower(%s, %d) iwconfig based", name, txpower)
		if out, err := core.Exec("iwconfig", []string{name, "txpower", fmt.Sprintf("%d", txpower)}); err != nil {
			return err
		} else if out != "" {
			return fmt.Errorf("unexpected output while setting txpower to %d for interface %s: %s", txpower, name, out)
		}
	}
	return nil
}

func GatewayProvidedByUser(iface *Endpoint, gateway string) (*Endpoint, error) {
	if IPv4Validator.MatchString(gateway) {
		Debug("valid gateway ip %s", gateway)
		// we have the address, now we need its mac
		if mac, err := ArpLookup(iface.Name(), gateway, false); err != nil {
			return nil, err
		} else {
			Debug("gateway is %s[%s]", gateway, mac)
			return NewEndpoint(gateway, mac), nil
		}
	}
	return nil, fmt.Errorf("Provided gateway %s not a valid IPv4 address! Revert to find default gateway.", gateway)
}

func ColorRSSI(n int) string {
	// ref. https://www.metageek.com/training/resources/understanding-rssi-2.html
	rssi := fmt.Sprintf("%d dBm", n)
	if n >= -67 {
		rssi = tui.Green(rssi)
	} else if n >= -70 {
		rssi = tui.Dim(tui.Green(rssi))
	} else if n >= -80 {
		rssi = tui.Yellow(rssi)
	} else {
		rssi = tui.Dim(tui.Red(rssi))
	}
	return rssi
}
