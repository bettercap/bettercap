package network

import "testing"

func buildExampleWiFi() *WiFi {
	return NewWiFi(buildExampleEndpoint(), func(ap *AccessPoint) {}, func(ap *AccessPoint) {})
}
