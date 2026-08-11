package ui

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/bettercap/bettercap/v2/network"
	"github.com/evilsocket/islazy/data"
)

func TestBundledWebUIWiFiAPIContract(t *testing.T) {
	javascript, err := web.ReadFile("ui/main.js")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(javascript, []byte("session.wifi['aps']")) {
		t.Fatal("bundled UI no longer reads the expected session.wifi.aps path")
	}

	apFields := []string{
		"alias", "authentication", "channel", "cipher", "clients", "encryption", "handshake",
		"hostname", "last_seen", "mac", "received", "rssi", "sent", "vendor", "wps",
	}
	clientFields := []string{
		"alias", "first_seen", "last_seen", "mac", "received", "rssi", "sent", "vendor",
	}
	for _, field := range apFields {
		if !bytes.Contains(javascript, []byte("ap."+field)) {
			t.Fatalf("bundled UI no longer references AP field %q", field)
		}
	}
	for _, field := range clientFields {
		if !bytes.Contains(javascript, []byte("client."+field)) {
			t.Fatalf("bundled UI no longer references client field %q", field)
		}
	}

	aliases, err := data.NewMemUnsortedKV()
	if err != nil {
		t.Fatal(err)
	}
	wifi := network.NewWiFi(nil, aliases, nil, nil)
	ap, _ := wifi.AddIfNew("test-network", "02:00:00:00:00:01", 2412, -42)
	ap.SetAlias("router")
	ap.SetEncryption("WPA2", "CCMP", "PSK")
	ap.SetWPS("Version", "2.0")
	ap.AddTraffic(10, 20)
	client, _ := ap.AddClientIfNew("02:00:00:00:00:02", 2412, -55)
	client.SetAlias("phone")
	client.SetVendor("Phone Vendor")
	client.AddTraffic(30, 40)
	ap.WithKeyMaterial(true)

	raw, err := json.Marshal(wifi)
	if err != nil {
		t.Fatal(err)
	}
	var response struct {
		APs []map[string]interface{} `json:"aps"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		t.Fatal(err)
	}
	if len(response.APs) != 1 {
		t.Fatalf("got %d APs, want 1", len(response.APs))
	}
	apJSON := response.APs[0]
	for _, field := range apFields {
		if _, found := apJSON[field]; !found {
			t.Fatalf("API no longer provides Web UI AP field %q: %#v", field, apJSON)
		}
	}
	if _, ok := apJSON["wps"].(map[string]interface{}); !ok {
		t.Fatalf("AP wps is no longer an object: %#v", apJSON["wps"])
	}
	if _, ok := apJSON["handshake"].(bool); !ok {
		t.Fatalf("AP handshake is no longer a boolean: %#v", apJSON["handshake"])
	}
	clients, ok := apJSON["clients"].([]interface{})
	if !ok || len(clients) != 1 {
		t.Fatalf("AP clients is no longer a populated array: %#v", apJSON["clients"])
	}
	clientJSON, ok := clients[0].(map[string]interface{})
	if !ok {
		t.Fatalf("client is no longer an object: %#v", clients[0])
	}
	for _, field := range clientFields {
		if _, found := clientJSON[field]; !found {
			t.Fatalf("API no longer provides Web UI client field %q: %#v", field, clientJSON)
		}
	}
}
