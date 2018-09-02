package network

import (
	"net"
	"testing"
)

func TestIsZeroMac(t *testing.T) {
	exampleMAC, _ := net.ParseMAC("00:00:00:00:00:00")

	exp := true
	got := IsZeroMac(exampleMAC)
	if got != exp {
		t.Fatalf("expected '%t', got '%t'", exp, got)
	}
}

func TestIsBroadcastMac(t *testing.T) {
	exampleMAC, _ := net.ParseMAC("ff:ff:ff:ff:ff:ff")

	exp := true
	got := IsBroadcastMac(exampleMAC)
	if got != exp {
		t.Fatalf("expected '%t', got '%t'", exp, got)
	}
}

func TestNormalizeMac(t *testing.T) {
	exp := "ff:ff:ff:ff:ff:ff"
	got := NormalizeMac("fF-fF-fF-fF-fF-fF")
	if got != exp {
		t.Fatalf("expected '%s', got '%s'", exp, got)
	}
}

// TODO: refactor to parse targets with an actual alias map
func TestParseTargets(t *testing.T) {
	cases := []struct {
		Name             string
		InputTargets     string
		InputAliases     *Aliases
		ExpectedIPCount  int
		ExpectedMACCount int
		ExpectedError    bool
	}{
		// Not sure how to trigger sad path where macParser.FindAllString()
		// finds a MAC but net.ParseMac() fails on the result.
		{
			"empty target string causes empty return",
			"",
			&Aliases{},
			0,
			0,
			false,
		},
		{
			"MACs are parsed",
			"192.168.1.2, 192.168.1.3, 5c:00:0b:90:a9:f0, 6c:00:0b:90:a9:f0",
			&Aliases{},
			2,
			2,
			false,
		},
	}
	for _, test := range cases {
		t.Run(test.Name, func(t *testing.T) {
			ips, macs, err := ParseTargets(test.InputTargets, test.InputAliases)
			if err != nil && !test.ExpectedError {
				t.Errorf("unexpected error: %s", err)
			}
			if err == nil && test.ExpectedError {
				t.Error("Expected error, but got none")
			}
			if test.ExpectedError {
				return
			}
			if len(ips) != test.ExpectedIPCount {
				t.Errorf("Wrong number of IPs. Got %v for targets %s", ips, test.InputTargets)
			}
			if len(macs) != test.ExpectedMACCount {
				t.Errorf("Wrong number of MACs. Got %v for targets %s", macs, test.InputTargets)
			}
		})
	}
}

func TestBuildEndpointFromInterface(t *testing.T) {
	ifaces, err := net.Interfaces()
	if err != nil {
		t.Error(err)
	}
	if len(ifaces) <= 0 {
		t.Error("Unable to find any network interfaces to run test with.")
	}
	_, err = buildEndpointFromInterface(ifaces[0])
	if err != nil {
		t.Error(err)
	}
}

func TestFindInterfaceByName(t *testing.T) {
	ifaces, err := net.Interfaces()
	if err != nil {
		t.Error(err)
	}
	if len(ifaces) <= 0 {
		t.Error("Unable to find any network interfaces to run test with.")
	}
	var exampleIface net.Interface
	// emulate libpcap's pcap_lookupdev function to find
	// default interface to test with ( maybe could use loopback ? )
	for _, iface := range ifaces {
		if iface.HardwareAddr != nil {
			exampleIface = iface
			break
		}
	}
	foundEndpoint, err := findInterfaceByName(exampleIface.Name, ifaces)
	if err != nil {
		t.Error("unable to find a given interface by name to build endpoint", err)
	}
	if foundEndpoint.Name() != exampleIface.Name {
		t.Error("unable to find a given interface by name to build endpoint")
	}
}

func TestFindInterface(t *testing.T) {
	ifaces, err := net.Interfaces()
	if err != nil {
		t.Error(err)
	}
	if len(ifaces) <= 0 {
		t.Error("Unable to find any network interfaces to run test with.")
	}
	var exampleIface net.Interface
	// emulate libpcap's pcap_lookupdev function to find
	// default interface to test with ( maybe could use loopback ? )
	for _, iface := range ifaces {
		if iface.HardwareAddr != nil {
			exampleIface = iface
			break
		}
	}
	foundEndpoint, err := FindInterface(exampleIface.Name)
	if err != nil {
		t.Error("unable to find a given interface by name to build endpoint", err)
	}
	if foundEndpoint.Name() != exampleIface.Name {
		t.Error("unable to find a given interface by name to build endpoint")
	}
}
