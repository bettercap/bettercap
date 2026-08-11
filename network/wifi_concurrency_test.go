package network

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/evilsocket/islazy/data"
	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
)

var stationJSONKeys = []string{
	"ipv4", "ipv6", "mac", "hostname", "alias", "vendor", "first_seen", "last_seen", "meta",
	"frequency", "channel", "rssi", "sent", "received", "encryption", "cipher", "authentication", "wps",
}

func decodeJSONObject(t *testing.T, raw []byte) map[string]interface{} {
	t.Helper()
	var doc map[string]interface{}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("invalid JSON %q: %v", raw, err)
	}
	return doc
}

func requireExactKeys(t *testing.T, doc map[string]interface{}, keys ...string) {
	t.Helper()
	want := make(map[string]bool, len(keys))
	for _, key := range keys {
		want[key] = true
	}
	if len(doc) != len(want) {
		t.Fatalf("unexpected key count: got %d (%v), want %d (%v)", len(doc), doc, len(want), keys)
	}
	for key := range doc {
		if !want[key] {
			t.Fatalf("unexpected JSON key %q in %v", key, doc)
		}
	}
}

func TestStationJSONContract(t *testing.T) {
	station := NewStation("test-network", "02:00:00:00:00:01", 2412, -42)
	station.SetAlias("router")
	station.SetVendor("Test Vendor")
	station.Snapshot().Meta.Set("source", "fixture")
	station.SetEncryption("WPA2", "CCMP", "PSK")
	station.SetWPS("Version", "2.0")
	station.AddTraffic(123, 456)

	raw, err := json.Marshal(station)
	if err != nil {
		t.Fatal(err)
	}
	doc := decodeJSONObject(t, raw)
	requireExactKeys(t, doc, stationJSONKeys...)

	wantValues := map[string]interface{}{
		"mac":            "02:00:00:00:00:01",
		"hostname":       "test-network",
		"alias":          "router",
		"vendor":         "Test Vendor",
		"frequency":      float64(2412),
		"channel":        float64(1),
		"rssi":           float64(-42),
		"sent":           float64(123),
		"received":       float64(456),
		"encryption":     "WPA2",
		"cipher":         "CCMP",
		"authentication": "PSK",
	}
	for key, want := range wantValues {
		if got := doc[key]; !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %#v, want %#v", key, got, want)
		}
	}
	if got := doc["wps"].(map[string]interface{})["Version"]; got != "2.0" {
		t.Fatalf("unexpected WPS data: %#v", doc["wps"])
	}

	var roundTrip Station
	if err := json.Unmarshal(raw, &roundTrip); err != nil {
		t.Fatal(err)
	}
	roundTripRaw, err := json.Marshal(&roundTrip)
	if err != nil {
		t.Fatal(err)
	}
	if got := decodeJSONObject(t, roundTripRaw); !reflect.DeepEqual(got, doc) {
		t.Fatalf("station JSON changed after round trip:\ngot  %#v\nwant %#v", got, doc)
	}
}

func TestAccessPointJSONContractAndRoundTrip(t *testing.T) {
	aliases, err := data.NewMemUnsortedKV()
	if err != nil {
		t.Fatal(err)
	}
	ap := NewAccessPoint("test-network", "02:00:00:00:00:01", 2412, -42, aliases)
	ap.Snapshot().Meta.Set("source", "fixture")
	ap.SetEncryption("WPA2", "CCMP", "PSK")
	ap.SetWPS("Version", "2.0")
	ap.AddTraffic(10, 20)
	client, added := ap.AddClientIfNew("02:00:00:00:00:02", 2412, -55)
	if !added {
		t.Fatal("expected a new client")
	}
	client.SetAlias("phone")
	client.Snapshot().Meta.Set("kind", "phone")
	client.AddTraffic(30, 40)
	ap.WithKeyMaterial(true)

	raw, err := json.Marshal(ap)
	if err != nil {
		t.Fatal(err)
	}
	doc := decodeJSONObject(t, raw)
	apKeys := append(append([]string(nil), stationJSONKeys...), "clients", "handshake")
	requireExactKeys(t, doc, apKeys...)
	if doc["handshake"] != true {
		t.Fatalf("handshake flag was not serialized: %#v", doc["handshake"])
	}
	clients, ok := doc["clients"].([]interface{})
	if !ok || len(clients) != 1 {
		t.Fatalf("unexpected clients: %#v", doc["clients"])
	}
	requireExactKeys(t, clients[0].(map[string]interface{}), stationJSONKeys...)

	var roundTrip AccessPoint
	if err := json.Unmarshal(raw, &roundTrip); err != nil {
		t.Fatal(err)
	}
	if !roundTrip.HasKeyMaterial() {
		t.Fatal("handshake flag was not restored during unmarshal")
	}
	roundTripRaw, err := json.Marshal(&roundTrip)
	if err != nil {
		t.Fatal(err)
	}
	if got := decodeJSONObject(t, roundTripRaw); !reflect.DeepEqual(got, doc) {
		t.Fatalf("AP JSON changed after round trip:\ngot  %#v\nwant %#v", got, doc)
	}
}

func TestZeroAccessPointJSONIsValid(t *testing.T) {
	raw, err := json.Marshal(&AccessPoint{})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(raw), `{"clients":[],"handshake":false}`; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestZeroStationJSONMatchesLegacyShape(t *testing.T) {
	const want = `{"frequency":0,"channel":0,"rssi":0,"sent":0,"received":0,"encryption":"","cipher":"","authentication":"","wps":null}`

	raw, err := json.Marshal(&Station{})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(raw); got != want {
		t.Fatalf("got %s, want %s", got, want)
	}

	var roundTrip Station
	if err := json.Unmarshal(raw, &roundTrip); err != nil {
		t.Fatal(err)
	}
	raw, err = json.Marshal(&roundTrip)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(raw); got != want {
		t.Fatalf("round trip: got %s, want %s", got, want)
	}
}

func TestEndpointlessAccessPointJSONMatchesLegacyShape(t *testing.T) {
	const fixture = `{"frequency":0,"channel":0,"rssi":0,"sent":0,"received":0,"encryption":"","cipher":"","authentication":"","wps":null,"clients":[{"frequency":0,"channel":0,"rssi":0,"sent":0,"received":0,"encryption":"","cipher":"","authentication":"","wps":null}],"handshake":false}`

	var ap AccessPoint
	if err := json.Unmarshal([]byte(fixture), &ap); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(&ap)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(raw); got != fixture {
		t.Fatalf("got %s, want %s", got, fixture)
	}
}

func TestNilWiFiEntitiesMarshalAsNull(t *testing.T) {
	for name, entity := range map[string]interface{}{
		"station":      (*Station)(nil),
		"access point": (*AccessPoint)(nil),
	} {
		raw, err := json.Marshal(entity)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got, want := string(raw), "null"; got != want {
			t.Fatalf("%s: got %s, want %s", name, got, want)
		}
	}
}

func TestWiFiJSONContract(t *testing.T) {
	aliases, err := data.NewMemUnsortedKV()
	if err != nil {
		t.Fatal(err)
	}
	wifi := NewWiFi(nil, aliases, nil, nil)
	wifi.AddIfNew("test-network", "02:00:00:00:00:01", 2412, -42)

	raw, err := json.Marshal(wifi)
	if err != nil {
		t.Fatal(err)
	}
	doc := decodeJSONObject(t, raw)
	requireExactKeys(t, doc, "aps")
	aps, ok := doc["aps"].([]interface{})
	if !ok || len(aps) != 1 {
		t.Fatalf("unexpected aps response: %#v", doc["aps"])
	}
	apKeys := append(append([]string(nil), stationJSONKeys...), "clients", "handshake")
	requireExactKeys(t, aps[0].(map[string]interface{}), apKeys...)
}

func TestEmptyWiFiJSONContract(t *testing.T) {
	aliases, err := data.NewMemUnsortedKV()
	if err != nil {
		t.Fatal(err)
	}
	wifi := NewWiFi(nil, aliases, nil, nil)

	raw, err := json.Marshal(wifi)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(raw), `{"aps":[]}`; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestStationConcurrentMutationAndSerialization(t *testing.T) {
	station := NewStation("test-network", "02:00:00:00:00:01", 2412, -42)
	packet := gopacket.NewPacket(make([]byte, 64), layers.LayerTypeEthernet, gopacket.Default)
	const iterations = 500

	errs := make(chan error, 8)
	var wg sync.WaitGroup
	run := func(fn func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := fn(); err != nil {
				errs <- err
			}
		}()
	}

	run(func() error {
		for i := 0; i < iterations; i++ {
			station.SetAlias(fmt.Sprintf("alias-%d", i))
			station.SetVendor(fmt.Sprintf("vendor-%d", i))
			station.AddTraffic(1, 2)
		}
		return nil
	})
	run(func() error {
		for i := 0; i < iterations; i++ {
			station.SetEncryption("WPA2", "CCMP", "PSK")
			station.SetWPS("Version", fmt.Sprint(i))
		}
		return nil
	})
	run(func() error {
		handshake := station.Handshake()
		for i := 0; i < iterations; i++ {
			handshake.UpdateBeacon(packet)
			handshake.AddFrame(i%3, packet)
		}
		return nil
	})
	for reader := 0; reader < 3; reader++ {
		run(func() error {
			for i := 0; i < iterations; i++ {
				station.Snapshot()
				_ = station.String()
				_ = station.PathFriendlyName()
				if _, err := json.Marshal(station); err != nil {
					return err
				}
				handshake := station.Handshake()
				handshake.Beacon()
				handshake.Complete()
			}
			return nil
		})
	}

	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	snapshot := station.Snapshot()
	if snapshot.Sent != iterations || snapshot.Received != iterations*2 {
		t.Fatalf("lost traffic updates: sent=%d received=%d", snapshot.Sent, snapshot.Received)
	}
}

func requireCompletes(t *testing.T, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("operation timed out; probable deadlock")
	}
}

func TestWiFiAndAccessPointCallbacksCanReenter(t *testing.T) {
	aliases, err := data.NewMemUnsortedKV()
	if err != nil {
		t.Fatal(err)
	}
	var wifi *WiFi
	wifi = NewWiFi(nil, aliases, func(ap *AccessPoint) {
		wifi.Get(ap.BSSID())
		wifi.List()
	}, func(ap *AccessPoint) {
		wifi.Get(ap.BSSID())
		wifi.List()
	})

	requireCompletes(t, func() {
		wifi.AddIfNew("test-network", "02:00:00:00:00:01", 2412, -42)
		wifi.Remove("02:00:00:00:00:01")
	})

	ap, _ := wifi.AddIfNew("test-network", "02:00:00:00:00:01", 2412, -42)
	ap.AddClientIfNew("02:00:00:00:00:02", 2412, -55)
	requireCompletes(t, func() {
		wifi.EachAccessPoint(func(mac string, current *AccessPoint) {
			wifi.Get(mac)
			wifi.Remove(mac)
		})
		ap.EachClient(func(mac string, station *Station) {
			ap.Get(mac)
			ap.RemoveClient(mac)
		})
	})
}

func TestWiFiConcurrentMutationAndMarshal(t *testing.T) {
	aliases, err := data.NewMemUnsortedKV()
	if err != nil {
		t.Fatal(err)
	}
	wifi := NewWiFi(nil, aliases, nil, nil)
	ap, _ := wifi.AddIfNew("test-network", "02:00:00:00:00:01", 2412, -42)
	client, _ := ap.AddClientIfNew("02:00:00:00:00:02", 2412, -55)

	const iterations = 400
	errs := make(chan error, 4)
	var wg sync.WaitGroup
	run := func(fn func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := fn(); err != nil {
				errs <- err
			}
		}()
	}

	run(func() error {
		for i := 0; i < iterations; i++ {
			wifi.AddIfNew(fmt.Sprintf("network-%d", i), ap.BSSID(), 2412, int8(-30-i%40))
			ap.AddClientIfNew(client.BSSID(), 2412, int8(-40-i%30))
			client.AddTraffic(1, 1)
		}
		return nil
	})
	run(func() error {
		for i := 0; i < iterations; i++ {
			mac := fmt.Sprintf("02:01:%02x:%02x:%02x:%02x", byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
			wifi.AddIfNew("temporary", mac, 2412, -60)
			wifi.Remove(mac)
		}
		return nil
	})
	for reader := 0; reader < 2; reader++ {
		run(func() error {
			for i := 0; i < iterations; i++ {
				if _, err := json.Marshal(wifi); err != nil {
					return err
				}
				if _, err := json.Marshal(ap); err != nil {
					return err
				}
				wifi.List()
				ap.Clients()
			}
			return nil
		})
	}

	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}

func TestHandshakeCallbackRunsUnlocked(t *testing.T) {
	handshake := NewHandshake()
	packet := gopacket.NewPacket(make([]byte, 64), layers.LayerTypeEthernet, gopacket.Default)
	handshake.UpdateBeacon(packet)
	handshake.AddExtra(packet)
	requireCompletes(t, func() {
		handshake.EachUnsavedPacket(func(gopacket.Packet) {
			handshake.Beacon()
			handshake.NumUnsaved()
		})
	})
}
