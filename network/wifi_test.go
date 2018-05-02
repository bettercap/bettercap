package network

import "testing"

func buildExampleWiFi() *WiFi {
	return NewWiFi(buildExampleEndpoint(), func(ap *AccessPoint) {}, func(ap *AccessPoint) {})
}

func TestDot11Freq2Chan(t *testing.T) {
	exampleFreq := 2472
	exp := 13
	got := Dot11Freq2Chan(exampleFreq)
	if got != exp {
		t.Fatalf("expected '%v', got '%v'", exp, got)
	}
}

func TestDot11Chan2Freq(t *testing.T) {
	exampleChan := 13
	exp := 2472
	got := Dot11Chan2Freq(exampleChan)
	if got != exp {
		t.Fatalf("expected '%v', got '%v'", exp, got)
	}
}
