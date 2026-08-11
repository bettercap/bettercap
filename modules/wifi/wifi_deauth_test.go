package wifi

import (
	"fmt"
	"net"
	"slices"
	"sort"
	"testing"

	"github.com/bettercap/bettercap/v2/network"
)

func deauthFlowPairs(flows []wifiTarget) []string {
	pairs := make([]string, 0, len(flows))
	for _, flow := range flows {
		pairs = append(pairs, fmt.Sprintf("%s/%s", flow.apSnapshot.HwAddress, flow.clientSnapshot.HwAddress))
	}
	sort.Strings(pairs)
	return pairs
}

func requireDeauthFlows(t *testing.T, mod *WiFiModule, target string, expected ...string) {
	t.Helper()
	selector, err := newWiFiSelector(target)
	if err != nil {
		t.Fatalf("selector %q: %v", target, err)
	}
	got := deauthFlowPairs(mod.deauthFlowsFor(selector))
	sort.Strings(expected)
	if !slices.Equal(got, expected) {
		t.Fatalf("selector %q: got %v, want %v", target, got, expected)
	}
}

func TestDeauthLegacySelectors(t *testing.T) {
	mod := newWiFiTargetTestModule(t)

	requireDeauthFlows(t, mod, "02:00:00:00:00:01",
		"02:00:00:00:00:01/02:00:00:00:01:01")
	requireDeauthFlows(t, mod, "02:00:00:00:02:02",
		"02:00:00:00:00:02/02:00:00:00:02:02")
	requireDeauthFlows(t, mod, "02:00:00:00:00:05")
	requireDeauthFlows(t, mod, "02:00:00:00:00:ff")

	all := []string{
		"02:00:00:00:00:01/02:00:00:00:01:01",
		"02:00:00:00:00:02/02:00:00:00:02:01",
		"02:00:00:00:00:02/02:00:00:00:02:02",
		"02:00:00:00:00:03/02:00:00:00:03:01",
		"02:00:00:00:00:04/02:00:00:00:04:01",
	}
	for _, target := range []string{"all", "*", network.BroadcastMac, "FF:FF:FF:FF:FF:FF"} {
		requireDeauthFlows(t, mod, target, append([]string(nil), all...)...)
	}
}

func TestDeauthExactESSIDSelectsEveryMatchingBSSID(t *testing.T) {
	mod := newWiFiTargetTestModule(t)
	requireDeauthFlows(t, mod, "Corp WiFi",
		"02:00:00:00:00:01/02:00:00:00:01:01",
		"02:00:00:00:00:02/02:00:00:00:02:01",
		"02:00:00:00:00:02/02:00:00:00:02:02")
}

func TestDeauthESSIDWildcardIsCaseSensitive(t *testing.T) {
	mod := newWiFiTargetTestModule(t)
	requireDeauthFlows(t, mod, "Corp*",
		"02:00:00:00:00:01/02:00:00:00:01:01",
		"02:00:00:00:00:02/02:00:00:00:02:01",
		"02:00:00:00:00:02/02:00:00:00:02:02")
	requireDeauthFlows(t, mod, "*IoT",
		"02:00:00:00:00:03/02:00:00:00:03:01")
}

func TestDeauthSkipListStillWinsForESSIDTargets(t *testing.T) {
	mod := newWiFiTargetTestModule(t)
	skippedAP, _ := net.ParseMAC("02:00:00:00:00:01")
	skippedClient, _ := net.ParseMAC("02:00:00:00:02:02")
	mod.deauthSkip = []net.HardwareAddr{skippedAP, skippedClient}

	requireDeauthFlows(t, mod, "Corp WiFi",
		"02:00:00:00:00:02/02:00:00:00:02:01")
	requireDeauthFlows(t, mod, "Corp*",
		"02:00:00:00:00:02/02:00:00:00:02:01")
	requireDeauthFlows(t, mod, "02:00:00:00:00:01")
	requireDeauthFlows(t, mod, "02:00:00:00:02:02")
}

func TestDeauthUnknownOrClientlessESSIDHasNoFlows(t *testing.T) {
	mod := newWiFiTargetTestModule(t)
	for _, target := range []string{"Missing", "Empty"} {
		selector, err := newWiFiSelector(target)
		if err != nil {
			t.Fatal(err)
		}
		if flows := mod.deauthFlowsFor(selector); len(flows) != 0 {
			t.Fatalf("selector %q unexpectedly resolved %d flows", target, len(flows))
		}
	}
}

func TestDeauthHandlerAcceptsESSIDsAndWildcards(t *testing.T) {
	mod := NewWiFiModule(createMockSession())
	for _, handler := range mod.Handlers() {
		if handler.Name != "wifi.deauth BSSID" {
			continue
		}
		for _, command := range []string{"wifi.deauth Corp WiFi", "wifi.deauth Corp*", "wifi.deauth all", "wifi.deauth 02:00:00:00:00:01"} {
			matched, args := handler.Parse(command)
			if !matched || len(args) != 1 || args[0] != command[len("wifi.deauth "):] {
				t.Fatalf("command %q parsed as matched=%v args=%q", command, matched, args)
			}
		}
		return
	}
	t.Fatal("wifi.deauth BSSID handler not found")
}

func TestDeauthCompleterIncludesUniqueESSIDsAndAddresses(t *testing.T) {
	mod := newWiFiTargetTestModule(t)
	got := mod.deauthCompleter("Corp")
	if !slices.Equal(got, []string{"", "Corp WiFi"}) {
		t.Fatalf("got %v, want [ Corp WiFi]", got)
	}

	got = mod.deauthCompleter("02:00:00:00:02")
	want := []string{"", "02:00:00:00:02:01", "02:00:00:00:02:02"}
	sort.Strings(got)
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
