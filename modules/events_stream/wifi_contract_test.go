package events_stream

import (
	"encoding/json"
	"testing"

	"github.com/bettercap/bettercap/v2/modules/wifi"
	"github.com/bettercap/bettercap/v2/network"
	"github.com/bettercap/bettercap/v2/session"
	"github.com/evilsocket/islazy/data"
)

func buildWiFiClientEvent(t *testing.T) session.Event {
	t.Helper()
	aliases, err := data.NewMemUnsortedKV()
	if err != nil {
		t.Fatal(err)
	}
	ap := network.NewAccessPoint("test-network", "02:00:00:00:00:01", 2412, -42, aliases)
	client, added := ap.AddClientIfNew("02:00:00:00:00:02", 2412, -55)
	if !added {
		t.Fatal("expected a new client")
	}
	return session.NewEvent("wifi.client.new", wifi.ClientEvent{AP: ap, Client: client})
}

func TestWiFiClientEventScriptingContract(t *testing.T) {
	event := buildWiFiClientEvent(t)
	raw, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}

	// JavaScript onEvent performs this exact JSON-to-generic-object conversion.
	var opaque map[string]interface{}
	if err := json.Unmarshal(raw, &opaque); err != nil {
		t.Fatal(err)
	}
	if opaque["tag"] != "wifi.client.new" {
		t.Fatalf("unexpected event tag: %#v", opaque["tag"])
	}
	eventData, ok := opaque["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected event data: %#v", opaque["data"])
	}
	if len(eventData) != 2 || eventData["AP"] == nil || eventData["Client"] == nil {
		t.Fatalf("AP/Client event keys changed: %#v", eventData)
	}
	ap := eventData["AP"].(map[string]interface{})
	client := eventData["Client"].(map[string]interface{})
	if ap["mac"] != "02:00:00:00:00:01" || client["mac"] != "02:00:00:00:00:02" {
		t.Fatalf("unexpected event addresses: AP=%#v Client=%#v", ap["mac"], client["mac"])
	}
	if _, ok := ap["clients"].([]interface{}); !ok {
		t.Fatalf("AP clients field changed: %#v", ap["clients"])
	}
}

func TestWiFiClientEventTriggerContract(t *testing.T) {
	for _, placeholder := range []string{
		`{{Client\mac}}`, // accepted by older built-in examples
		`{{Client/mac}}`, // documented caplet syntax
		`{{/Client/mac}}`,
	} {
		t.Run(placeholder, func(t *testing.T) {
			triggers := NewTriggerList()
			if err, _ := triggers.Add("wifi.client.new", "wifi.deauth "+placeholder); err != nil {
				t.Fatal(err)
			}

			_, command, err, found := triggers.Dispatch(buildWiFiClientEvent(t))
			if err != nil {
				t.Fatal(err)
			}
			if !found {
				t.Fatal("trigger was not dispatched")
			}
			if want := "wifi.deauth 02:00:00:00:00:02"; command != want {
				t.Fatalf("got command %q, want %q", command, want)
			}
		})
	}
}
