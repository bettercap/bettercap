package wifi

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bettercap/bettercap/v2/session"
)

func addWiFiTargetTestAP(t *testing.T, mod *WiFiModule, essid, bssid string, clients ...string) {
	t.Helper()
	ap, added := mod.Session.WiFi.AddIfNew(essid, bssid, 2412, -42)
	if !added {
		t.Fatalf("AP %s was not added", bssid)
	}
	for _, client := range clients {
		if _, added := ap.AddClientIfNew(client, 2412, -55); !added {
			t.Fatalf("client %s was not added to %s", client, bssid)
		}
	}
}

func newWiFiTargetTestModule(t *testing.T) *WiFiModule {
	t.Helper()
	mod := NewWiFiModule(createMockSession())
	addWiFiTargetTestAP(t, mod, "Corp WiFi", "02:00:00:00:00:01", "02:00:00:00:01:01")
	addWiFiTargetTestAP(t, mod, "Corp WiFi", "02:00:00:00:00:02", "02:00:00:00:02:01", "02:00:00:00:02:02")
	addWiFiTargetTestAP(t, mod, "CORP-IoT", "02:00:00:00:00:03", "02:00:00:00:03:01")
	addWiFiTargetTestAP(t, mod, "Guest", "02:00:00:00:00:04", "02:00:00:00:04:01")
	addWiFiTargetTestAP(t, mod, "Empty", "02:00:00:00:00:05")
	return mod
}

func newWiFiExpressionTestModule(t *testing.T) *WiFiModule {
	t.Helper()
	mod := NewWiFiModule(createMockSession())
	addWiFiTargetTestAP(t, mod, "Corp WiFi", "02:10:00:00:00:01", "02:10:00:00:01:01")
	addWiFiTargetTestAP(t, mod, "Corp WiFi", "02:10:00:00:00:02", "02:10:00:00:02:01", "02:10:00:00:02:02")
	addWiFiTargetTestAP(t, mod, "Corp1", "02:10:00:00:00:03", "02:10:00:00:03:01")
	addWiFiTargetTestAP(t, mod, "Corp2", "02:10:00:00:00:04", "02:10:00:00:04:01")
	addWiFiTargetTestAP(t, mod, "CorpA", "02:10:00:00:00:05", "02:10:00:00:05:01")
	addWiFiTargetTestAP(t, mod, "myguestnet", "02:10:00:00:00:06", "02:10:00:00:06:01")
	addWiFiTargetTestAP(t, mod, "MyGuestNet", "02:10:00:00:00:07", "02:10:00:00:07:01")
	addWiFiTargetTestAP(t, mod, "Office1", "02:10:00:00:00:08", "02:10:00:00:08:01")
	addWiFiTargetTestAP(t, mod, "Office12", "02:10:00:00:00:09", "02:10:00:00:09:01")
	addWiFiTargetTestAP(t, mod, "Empty", "02:10:00:00:00:0a")
	return mod
}

func TestWiFiSelectorRejectsInvalidWildcard(t *testing.T) {
	if _, err := newWiFiSelector("Corp["); err == nil {
		t.Fatal("expected an invalid wildcard error")
	}
}

func TestWiFiSelectorKeepsLegacyMACPrecedence(t *testing.T) {
	mod := NewWiFiModule(createMockSession())
	addWiFiTargetTestAP(t, mod, "Normal", "02:00:00:00:00:01", "02:00:00:00:01:01")
	addWiFiTargetTestAP(t, mod, "02:00:00:00:00:01", "02:00:00:00:00:02", "02:00:00:00:02:01")
	selector, err := newWiFiSelector("02:00:00:00:00:01")
	if err != nil {
		t.Fatal(err)
	}
	assocTargets := mod.assocTargetsFor(selector)
	if len(assocTargets) != 1 {
		t.Fatalf("assoc MAC selector resolved %d targets, want 1", len(assocTargets))
	}
	if got := assocTargets[0].apSnapshot.HwAddress; got != "02:00:00:00:00:01" {
		t.Fatalf("assoc MAC selector resolved %s, want 02:00:00:00:00:01", got)
	}
	deauthTargets := mod.deauthFlowsFor(selector)
	if got, want := deauthFlowPairs(deauthTargets), []string{"02:00:00:00:00:01/02:00:00:00:01:01"}; !slices.Equal(got, want) {
		t.Fatalf("deauth MAC selector resolved %v, want %v", got, want)
	}
}

func TestWiFiTargetExpressionMatrix(t *testing.T) {
	mod := newWiFiExpressionTestModule(t)
	tests := []struct {
		name       string
		selector   string
		assocAPs   []string
		deauthFlow []string
	}{
		{
			name:     "exact ESSID matches every BSSID",
			selector: "Corp WiFi",
			assocAPs: []string{"02:10:00:00:00:01", "02:10:00:00:00:02"},
			deauthFlow: []string{
				"02:10:00:00:00:01/02:10:00:00:01:01",
				"02:10:00:00:00:02/02:10:00:00:02:01",
				"02:10:00:00:00:02/02:10:00:00:02:02",
			},
		},
		{
			name:     "star matches an ESSID prefix",
			selector: "Corp*",
			assocAPs: []string{
				"02:10:00:00:00:01", "02:10:00:00:00:02", "02:10:00:00:00:03",
				"02:10:00:00:00:04", "02:10:00:00:00:05",
			},
			deauthFlow: []string{
				"02:10:00:00:00:01/02:10:00:00:01:01",
				"02:10:00:00:00:02/02:10:00:00:02:01",
				"02:10:00:00:00:02/02:10:00:00:02:02",
				"02:10:00:00:00:03/02:10:00:00:03:01",
				"02:10:00:00:00:04/02:10:00:00:04:01",
				"02:10:00:00:00:05/02:10:00:00:05:01",
			},
		},
		{
			name:       "star matching is case sensitive",
			selector:   "*guest*",
			assocAPs:   []string{"02:10:00:00:00:06"},
			deauthFlow: []string{"02:10:00:00:00:06/02:10:00:00:06:01"},
		},
		{
			name:       "question mark matches one character",
			selector:   "Office?",
			assocAPs:   []string{"02:10:00:00:00:08"},
			deauthFlow: []string{"02:10:00:00:00:08/02:10:00:00:08:01"},
		},
		{
			name:       "character class matches listed characters",
			selector:   "Corp[12]",
			assocAPs:   []string{"02:10:00:00:00:03", "02:10:00:00:00:04"},
			deauthFlow: []string{"02:10:00:00:00:03/02:10:00:00:03:01", "02:10:00:00:00:04/02:10:00:00:04:01"},
		},
		{
			name:     "unknown expression has no targets",
			selector: "Missing*",
		},
		{
			name:     "clientless AP is assoc-only",
			selector: "Empty",
			assocAPs: []string{"02:10:00:00:00:0a"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			selector, err := newWiFiSelector(test.selector)
			if err != nil {
				t.Fatal(err)
			}
			gotAssoc := assocTargetBSSIDs(mod.assocTargetsFor(selector))
			gotDeauth := deauthFlowPairs(mod.deauthFlowsFor(selector))
			sort.Strings(test.assocAPs)
			sort.Strings(test.deauthFlow)
			if !slices.Equal(gotAssoc, test.assocAPs) {
				t.Errorf("assoc target %q: got %v, want %v", test.selector, gotAssoc, test.assocAPs)
			}
			if !slices.Equal(gotDeauth, test.deauthFlow) {
				t.Errorf("deauth target %q: got %v, want %v", test.selector, gotDeauth, test.deauthFlow)
			}
		})
	}
}

func TestWiFiTargetHandlersRejectInvalidExpressions(t *testing.T) {
	mod := NewWiFiModule(createMockSession())
	for _, handler := range mod.Handlers() {
		if handler.Name != "wifi.deauth BSSID" && handler.Name != "wifi.assoc BSSID" {
			continue
		}
		err := handler.Exec([]string{"Corp["})
		if err == nil || !strings.Contains(err.Error(), "invalid ESSID expression") {
			t.Errorf("%s returned %v for invalid expression", handler.Name, err)
		}
	}
}

func TestWiFiTargetCommandResultsForEmptySelections(t *testing.T) {
	mod := newWiFiExpressionTestModule(t)
	mod.StatusLock.Lock()
	mod.Started = true
	mod.StatusLock.Unlock()

	if err := mod.startDeauth("Missing*"); err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("deauth unknown expression returned %v", err)
	}
	if err := mod.startAssoc("Missing*"); err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("assoc unknown expression returned %v", err)
	}
	if err := mod.startDeauth("Empty"); err == nil || !strings.Contains(err.Error(), "doesn't have detected clients") {
		t.Fatalf("deauth clientless ESSID returned %v", err)
	}
	if err := mod.startAssoc("Empty"); err != nil {
		t.Fatalf("assoc clientless ESSID returned %v", err)
	}
	mod.writes.Wait()

	empty := NewWiFiModule(createMockSession())
	empty.StatusLock.Lock()
	empty.Started = true
	empty.StatusLock.Unlock()
	if err := empty.startDeauth("all"); err != nil {
		t.Fatalf("deauth all on an empty scan returned %v", err)
	}
	if err := empty.startAssoc("*"); err != nil {
		t.Fatalf("assoc * on an empty scan returned %v", err)
	}
}

func TestLegacyWiFiTargetCommandSyntaxStillParses(t *testing.T) {
	mod := NewWiFiModule(createMockSession())
	tests := []struct {
		handlerName string
		command     string
		want        string
	}{
		{"wifi.deauth BSSID", "wifi.deauth 02:00:00:00:00:01", "02:00:00:00:00:01"},
		{"wifi.deauth BSSID", "wifi.deauth 02:00:00:00:01:01", "02:00:00:00:01:01"},
		{"wifi.deauth BSSID", "wifi.deauth all", "all"},
		{"wifi.deauth BSSID", "wifi.deauth *", "*"},
		{"wifi.deauth BSSID", "wifi.deauth ff:ff:ff:ff:ff:ff", "ff:ff:ff:ff:ff:ff"},
		{"wifi.deauth BSSID", "wifi.deauth FF:FF:FF:FF:FF:FF", "FF:FF:FF:FF:FF:FF"},
		{"wifi.assoc BSSID", "wifi.assoc 02:00:00:00:00:01", "02:00:00:00:00:01"},
		{"wifi.assoc BSSID", "wifi.assoc all", "all"},
		{"wifi.assoc BSSID", "wifi.assoc *", "*"},
		{"wifi.assoc BSSID", "wifi.assoc ff:ff:ff:ff:ff:ff", "ff:ff:ff:ff:ff:ff"},
		{"wifi.assoc BSSID", "wifi.assoc FF:FF:FF:FF:FF:FF", "FF:FF:FF:FF:FF:FF"},
	}

	handlers := make(map[string]session.ModuleHandler)
	for _, handler := range mod.Handlers() {
		handlers[handler.Name] = handler
	}
	for _, test := range tests {
		handler, found := handlers[test.handlerName]
		if !found {
			t.Fatalf("handler %q not found", test.handlerName)
		}
		matched, args := handler.Parse(test.command)
		if !matched || len(args) != 1 || args[0] != test.want {
			t.Errorf("%q parsed as matched=%v args=%q, want [%q]", test.command, matched, args, test.want)
		}
		if _, err := newWiFiSelector(test.want); err != nil {
			t.Errorf("legacy target %q is no longer valid: %v", test.want, err)
		}
	}
}

func TestWiFiTargetResolutionDoesNotDeadlockDuringReconUpdates(t *testing.T) {
	mod := newWiFiTargetTestModule(t)
	ap, found := mod.Session.WiFi.Get("02:00:00:00:00:01")
	if !found {
		t.Fatal("test AP not found")
	}
	selector, err := newWiFiSelector("Corp*")
	if err != nil {
		t.Fatal(err)
	}

	var workers sync.WaitGroup
	workers.Add(3)
	go func() {
		defer workers.Done()
		for i := 0; i < 250; i++ {
			mod.Session.WiFi.AddIfNew("Corp WiFi", "02:00:00:00:00:01", 2412, int8(-30-i%40))
			ap.AddClientIfNew(fmt.Sprintf("02:00:00:01:%02x:%02x", i/256, i%256), 2412, -55)
		}
	}()
	go func() {
		defer workers.Done()
		for i := 0; i < 250; i++ {
			_ = mod.deauthFlowsFor(selector)
			_ = mod.assocTargetsFor(selector)
		}
	}()
	go func() {
		defer workers.Done()
		for i := 0; i < 250; i++ {
			_ = mod.deauthCompleter("Corp")
			_ = mod.assocCompleter("Corp")
		}
	}()

	done := make(chan struct{})
	go func() {
		workers.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Wi-Fi target resolution deadlocked during concurrent recon updates")
	}
}
