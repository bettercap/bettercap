package wifi

import (
	"strings"
	"testing"
	"time"
)

func TestWiFiShowRowContract(t *testing.T) {
	sess := createMockSession()
	sess.StartedAt = time.Now().Add(-time.Hour)
	mod := NewWiFiModule(sess)
	mod.frequencies = []int{2412}
	mod.minRSSI = -100
	mod.showManuf = false

	ap, _ := sess.WiFi.AddIfNew("test-network", "02:00:00:00:00:01", 2412, -42)
	ap.SetEncryption("WPA2", "CCMP", "PSK")
	ap.SetWPS("Version", "2.0")
	ap.AddTraffic(1024, 2048)

	row, include := mod.getRow(ap.Station())
	if !include {
		t.Fatal("expected AP to be included")
	}
	if len(row) != 10 {
		t.Fatalf("got %d columns, want 10: %#v", len(row), row)
	}
	checks := map[int]string{
		1: "02:00:00:00:00:01",
		2: "test-network",
		3: "WPA2 (CCMP, PSK)",
		4: "2.0",
		5: "1",
		7: "1.0 kB",
		8: "2.0 kB",
	}
	for column, want := range checks {
		if !strings.Contains(row[column], want) {
			t.Errorf("column %d: got %q, want it to contain %q", column, row[column], want)
		}
	}
}

func TestWiFiSelectionFilterAndSortContract(t *testing.T) {
	sess := createMockSession()
	mod := NewWiFiModule(sess)

	first, _ := sess.WiFi.AddIfNew("first", "02:00:00:00:00:01", 2412, -60)
	first.SetEncryption("WPA2", "CCMP", "PSK")
	first.AddTraffic(20, 0)
	second, _ := sess.WiFi.AddIfNew("second", "02:00:00:00:00:02", 2412, -40)
	second.SetEncryption("WPA2", "CCMP", "PSK")
	second.AddTraffic(10, 0)
	open, _ := sess.WiFi.AddIfNew("open", "02:00:00:00:00:03", 2412, -20)
	open.SetEncryption("OPEN", "", "")

	sess.Env.Set("wifi.show.filter", "WPA2")
	sess.Env.Set("wifi.show.sort", "sent asc")
	sess.Env.Set("wifi.show.limit", "0")

	err, stations := mod.doSelection()
	if err != nil {
		t.Fatal(err)
	}
	if len(stations) != 2 {
		t.Fatalf("got %d filtered stations, want 2", len(stations))
	}
	if got := []string{stations[0].BSSID(), stations[1].BSSID()}; got[0] != second.BSSID() || got[1] != first.BSSID() {
		t.Fatalf("unexpected sent sort order: %#v", got)
	}
}

func TestWiFiRSSISortContract(t *testing.T) {
	sess := createMockSession()
	mod := NewWiFiModule(sess)

	weak, _ := sess.WiFi.AddIfNew("weak", "02:00:00:00:00:01", 2412, -84)
	mid, _ := sess.WiFi.AddIfNew("mid", "02:00:00:00:00:02", 2412, -64)
	strong, _ := sess.WiFi.AddIfNew("strong", "02:00:00:00:00:03", 2412, -47)

	sess.Env.Set("wifi.show.filter", "")
	sess.Env.Set("wifi.show.limit", "0")

	sess.Env.Set("wifi.show.sort", "rssi desc")
	err, stations := mod.doSelection()
	if err != nil {
		t.Fatal(err)
	}
	if got := []string{stations[0].BSSID(), stations[1].BSSID(), stations[2].BSSID()}; got[0] != strong.BSSID() || got[1] != mid.BSSID() || got[2] != weak.BSSID() {
		t.Fatalf("unexpected rssi desc sort order (want strongest signal first): %#v", got)
	}

	sess.Env.Set("wifi.show.sort", "rssi asc")
	err, stations = mod.doSelection()
	if err != nil {
		t.Fatal(err)
	}
	if got := []string{stations[0].BSSID(), stations[1].BSSID(), stations[2].BSSID()}; got[0] != weak.BSSID() || got[1] != mid.BSSID() || got[2] != strong.BSSID() {
		t.Fatalf("unexpected rssi asc sort order (want weakest signal first): %#v", got)
	}
}
