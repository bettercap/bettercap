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
