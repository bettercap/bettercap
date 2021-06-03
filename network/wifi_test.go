package network

import (
	"testing"

	"github.com/evilsocket/islazy/data"
)

// Define test data for dot11 frequency <-> channel tests
type dot11pair struct {
	frequency int
	channel   int
}

var dot11TestVector = []dot11pair{
	{2472, 13},
	{2484, 14},
	{5825, 165},
	{5885, 177},
}

func buildExampleWiFi() *WiFi {
	aliases := &data.UnsortedKV{}
	return NewWiFi(buildExampleEndpoint(), aliases, func(ap *AccessPoint) {}, func(ap *AccessPoint) {})
}

func TestDot11Freq2Chan(t *testing.T) {
	for _, entry := range dot11TestVector {
		gotChannel := Dot11Freq2Chan(entry.frequency)
		if gotChannel != entry.channel {
			t.Fatalf("expected '%v', got '%v'", entry.channel, gotChannel)
		}
	}
}

func TestDot11Chan2Freq(t *testing.T) {
	for _, entry := range dot11TestVector {
		gotFrequency := Dot11Chan2Freq(entry.channel)
		if gotFrequency != entry.frequency {
			t.Fatalf("expected '%v', got '%v'", entry.frequency, gotFrequency)
		}
	}
}

func TestNewWiFi(t *testing.T) {
	aliases := &data.UnsortedKV{}
	exampleWiFi := NewWiFi(buildExampleEndpoint(), aliases, func(ap *AccessPoint) {}, func(ap *AccessPoint) {})
	if exampleWiFi == nil {
		t.Error("unable to build net wifi struct")
	}
}

func TestWiFiMarshalJSON(t *testing.T) {
	exampleWiFi := buildExampleWiFi()
	json, err := exampleWiFi.MarshalJSON()
	if err != nil {
		t.Error(err)
	}
	if len(json) <= 0 {
		t.Error("unable to marshal JSON from WiFi struct")
	}
}

func TestEachAccessPoint(t *testing.T) {
	exampleWiFi := buildExampleWiFi()
	aliases := &data.UnsortedKV{}
	exampleAP := NewAccessPoint("my_wifi", "ff:ff:ff:ff:ff:ff", 2472, int8(0), aliases)
	exampleWiFi.aps["ff:ff:ff:ff:ff:f1"] = exampleAP
	exampleWiFi.aps["ff:ff:ff:ff:ff:f2"] = exampleAP
	count := 0
	exampleCB := func(mac string, ap *AccessPoint) {
		count++
	}
	exampleWiFi.EachAccessPoint(exampleCB)
	if count != 2 {
		t.Error("unable to perform callback function for each access point")
	}
}

func TestStations(t *testing.T) {
	exampleWiFi := buildExampleWiFi()
	aliases := &data.UnsortedKV{}
	exampleAP := NewAccessPoint("my_wifi", "ff:ff:ff:ff:ff:ff", 2472, int8(0), aliases)
	exampleWiFi.aps["ff:ff:ff:ff:ff:f1"] = exampleAP
	exampleWiFi.aps["ff:ff:ff:ff:ff:f2"] = exampleAP
	exp := 2
	got := len(exampleWiFi.Stations())
	if got != exp {
		t.Fatalf("expected '%v', got '%v'", exp, got)
	}
}

func TestWiFiList(t *testing.T) {
	exampleWiFi := buildExampleWiFi()
	aliases := &data.UnsortedKV{}
	exampleAP := NewAccessPoint("my_wifi", "ff:ff:ff:ff:ff:ff", 2472, int8(0), aliases)
	exampleWiFi.aps["ff:ff:ff:ff:ff:f1"] = exampleAP
	exampleWiFi.aps["ff:ff:ff:ff:ff:f2"] = exampleAP
	exp := 2
	got := len(exampleWiFi.List())
	if got != exp {
		t.Fatalf("expected '%v', got '%v'", exp, got)
	}
}

func TestWiFiRemove(t *testing.T) {
	exampleWiFi := buildExampleWiFi()
	aliases := &data.UnsortedKV{}
	exampleAP := NewAccessPoint("my_wifi", "ff:ff:ff:ff:ff:ff", 2472, int8(0), aliases)
	exampleWiFi.aps["ff:ff:ff:ff:ff:f1"] = exampleAP
	exampleWiFi.aps["ff:ff:ff:ff:ff:f2"] = exampleAP
	exampleWiFi.Remove("ff:ff:ff:ff:ff:f1")
	exp := 1
	got := len(exampleWiFi.List())
	if got != exp {
		t.Fatalf("expected '%v', got '%v'", exp, got)
	}
}

func TestWiFiAddIfNew(t *testing.T) {
	exampleWiFi := buildExampleWiFi()
	aliases := &data.UnsortedKV{}
	exampleAP := NewAccessPoint("my_wifi", "ff:ff:ff:ff:ff:ff", 2472, int8(0), aliases)
	exampleWiFi.aps["ff:ff:ff:ff:ff:f1"] = exampleAP
	exampleWiFi.aps["ff:ff:ff:ff:ff:f2"] = exampleAP
	exampleWiFi.AddIfNew("my_wifi2", "ff:ff:ff:ff:ff:f3", 2472, int8(0))
	exp := 3
	got := len(exampleWiFi.List())
	if got != exp {
		t.Fatalf("expected '%v', got '%v'", exp, got)
	}
}

func TestWiFiGet(t *testing.T) {
	exampleWiFi := buildExampleWiFi()
	aliases := &data.UnsortedKV{}
	exampleAP := NewAccessPoint("my_wifi", "ff:ff:ff:ff:ff:ff", 2472, int8(0), aliases)
	exampleWiFi.aps["ff:ff:ff:ff:ff:ff"] = exampleAP
	_, found := exampleWiFi.Get("ff:ff:ff:ff:ff:ff")
	if !found {
		t.Error("unable to get access point from wifi struct with mac address")
	}
}

func TestWiFiGetClient(t *testing.T) {
	exampleWiFi := buildExampleWiFi()
	aliases := &data.UnsortedKV{}
	exampleAP := NewAccessPoint("my_wifi", "ff:ff:ff:ff:ff:ff", 2472, int8(0), aliases)
	exampleClient := NewStation("my_wifi", "ff:ff:ff:ff:ff:xx", 2472, int8(0))
	exampleAP.clients["ff:ff:ff:ff:ff:xx"] = exampleClient
	exampleWiFi.aps["ff:ff:ff:ff:ff:ff"] = exampleAP
	_, found := exampleWiFi.GetClient("ff:ff:ff:ff:ff:xx")
	if !found {
		t.Error("unable to get client from wifi struct with mac address")
	}
}

func TestWiFiClear(t *testing.T) {
	exampleWiFi := buildExampleWiFi()
	aliases := &data.UnsortedKV{}
	exampleAP := NewAccessPoint("my_wifi", "ff:ff:ff:ff:ff:ff", 2472, int8(0), aliases)
	exampleWiFi.aps["ff:ff:ff:ff:ff:ff"] = exampleAP
	exampleWiFi.Clear()
	if len(exampleWiFi.aps) != 0 {
		t.Error("unable to clear known access point for wifi struct")
	}
}
