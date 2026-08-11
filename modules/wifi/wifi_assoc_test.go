package wifi

import (
	"net"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/bettercap/bettercap/v2/network"
)

func assocTargetBSSIDs(targets []wifiTarget) []string {
	bssids := make([]string, 0, len(targets))
	for _, target := range targets {
		bssids = append(bssids, target.apSnapshot.HwAddress)
	}
	sort.Strings(bssids)
	return bssids
}

func requireAssocTargets(t *testing.T, mod *WiFiModule, target string, expected ...string) {
	t.Helper()
	selector, err := newWiFiSelector(target)
	if err != nil {
		t.Fatalf("selector %q: %v", target, err)
	}
	got := assocTargetBSSIDs(mod.assocTargetsFor(selector))
	sort.Strings(expected)
	if !slices.Equal(got, expected) {
		t.Fatalf("selector %q: got %v, want %v", target, got, expected)
	}
}

func TestAssocLegacySelectors(t *testing.T) {
	mod := newWiFiTargetTestModule(t)

	requireAssocTargets(t, mod, "02:00:00:00:00:01", "02:00:00:00:00:01")
	requireAssocTargets(t, mod, "02:00:00:00:00:05", "02:00:00:00:00:05")
	requireAssocTargets(t, mod, "02:00:00:00:00:ff")

	all := []string{
		"02:00:00:00:00:01",
		"02:00:00:00:00:02",
		"02:00:00:00:00:03",
		"02:00:00:00:00:04",
		"02:00:00:00:00:05",
	}
	for _, target := range []string{"all", "*", network.BroadcastMac, "FF:FF:FF:FF:FF:FF"} {
		requireAssocTargets(t, mod, target, append([]string(nil), all...)...)
	}
}

func TestAssocExactESSIDSelectsEveryMatchingBSSID(t *testing.T) {
	mod := newWiFiTargetTestModule(t)
	requireAssocTargets(t, mod, "Corp WiFi",
		"02:00:00:00:00:01",
		"02:00:00:00:00:02")
}

func TestAssocESSIDWildcardIsCaseSensitive(t *testing.T) {
	mod := newWiFiTargetTestModule(t)
	requireAssocTargets(t, mod, "Corp*",
		"02:00:00:00:00:01",
		"02:00:00:00:00:02")
	requireAssocTargets(t, mod, "*IoT", "02:00:00:00:00:03")
}

func TestAssocESSIDCanSelectAPWithoutClients(t *testing.T) {
	mod := newWiFiTargetTestModule(t)
	requireAssocTargets(t, mod, "Empty", "02:00:00:00:00:05")
}

func TestAssocSkipListStillWinsForESSIDTargets(t *testing.T) {
	mod := newWiFiTargetTestModule(t)
	skippedAP, _ := net.ParseMAC("02:00:00:00:00:01")
	mod.assocSkip = []net.HardwareAddr{skippedAP}

	requireAssocTargets(t, mod, "Corp WiFi", "02:00:00:00:00:02")
	requireAssocTargets(t, mod, "Corp*", "02:00:00:00:00:02")
	requireAssocTargets(t, mod, "02:00:00:00:00:01")
}

func TestAssocUnknownESSIDHasNoTargets(t *testing.T) {
	mod := newWiFiTargetTestModule(t)
	requireAssocTargets(t, mod, "Missing")
}

func TestAssocHandlerAcceptsESSIDsAndWildcards(t *testing.T) {
	mod := NewWiFiModule(createMockSession())
	for _, handler := range mod.Handlers() {
		if handler.Name != "wifi.assoc BSSID" {
			continue
		}
		for _, command := range []string{"wifi.assoc Corp WiFi", "wifi.assoc Corp*", "wifi.assoc all", "wifi.assoc 02:00:00:00:00:01"} {
			matched, args := handler.Parse(command)
			if !matched || len(args) != 1 || args[0] != command[len("wifi.assoc "):] {
				t.Fatalf("command %q parsed as matched=%v args=%q", command, matched, args)
			}
		}
		return
	}
	t.Fatal("wifi.assoc BSSID handler not found")
}

func TestAssocCompleterIncludesUniqueESSIDsAndAPAddresses(t *testing.T) {
	mod := newWiFiTargetTestModule(t)
	got := mod.assocCompleter("Corp")
	if !slices.Equal(got, []string{"", "Corp WiFi"}) {
		t.Fatalf("got %v, want [ Corp WiFi]", got)
	}

	got = mod.assocCompleter("02:00:00:00:02")
	if !slices.Equal(got, []string{""}) {
		t.Fatalf("client BSSIDs must not be association completions: got %v", got)
	}

	got = mod.assocCompleter("02:00:00:00:00:02")
	want := []string{"", "02:00:00:00:00:02"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAssocTargetDescriptionIncludesESSIDSupport(t *testing.T) {
	mod := NewWiFiModule(createMockSession())
	for _, handler := range mod.Handlers() {
		if handler.Name == "wifi.assoc BSSID" {
			for _, expected := range []string{"exact ESSID", "wildcard", "every visible AP"} {
				if !strings.Contains(handler.Description, expected) {
					t.Fatalf("description %q does not contain %q", handler.Description, expected)
				}
			}
			return
		}
	}
	t.Fatal("wifi.assoc BSSID handler not found")
}
