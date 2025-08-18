package tcp_proxy

import (
	"net"
	"testing"

	"github.com/evilsocket/islazy/plugin"
)

func TestOnData_NoReturn(t *testing.T) {
	jsCode := `
		function onData(from, to, data, callback) {
			// don't return anything
		}
	`

	plug, err := plugin.Parse(jsCode)
	if err != nil {
		t.Fatalf("Failed to parse plugin: %v", err)
	}

	script := &TcpProxyScript{
		Plugin:   plug,
		doOnData: plug.HasFunc("onData"),
	}

	from := &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 1234}
	to := &net.TCPAddr{IP: net.ParseIP("192.168.1.2"), Port: 5678}
	data := []byte("test data")

	result := script.OnData(from, to, data, nil)
	if result != nil {
		t.Errorf("Expected nil result when callback returns nothing, got %v", result)
	}
}

func TestOnData_ReturnsArrayOfIntegers(t *testing.T) {
	jsCode := `
		function onData(from, to, data, callback) {
			// Return modified data as array of integers
			return [72, 101, 108, 108, 111]; // "Hello" in ASCII
		}
	`

	plug, err := plugin.Parse(jsCode)
	if err != nil {
		t.Fatalf("Failed to parse plugin: %v", err)
	}

	script := &TcpProxyScript{
		Plugin:   plug,
		doOnData: plug.HasFunc("onData"),
	}

	from := &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 1234}
	to := &net.TCPAddr{IP: net.ParseIP("192.168.1.2"), Port: 5678}
	data := []byte("test data")

	result := script.OnData(from, to, data, nil)
	expected := []byte("Hello")

	if result == nil {
		t.Fatal("Expected non-nil result when callback returns array of integers")
	}

	if len(result) != len(expected) {
		t.Fatalf("Expected result length %d, got %d", len(expected), len(result))
	}

	for i, b := range result {
		if b != expected[i] {
			t.Errorf("Expected byte at index %d to be %d, got %d", i, expected[i], b)
		}
	}
}

func TestOnData_ReturnsDynamicArray(t *testing.T) {
	jsCode := `
		function onData(from, to, data, callback) {
			var result = [];
			for (var i = 0; i < data.length; i++) {
				result.push((data[i] + 1) % 256);
			}
			return result;
		}
	`

	plug, err := plugin.Parse(jsCode)
	if err != nil {
		t.Fatalf("Failed to parse plugin: %v", err)
	}

	script := &TcpProxyScript{
		Plugin:   plug,
		doOnData: plug.HasFunc("onData"),
	}

	from := &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 1234}
	to := &net.TCPAddr{IP: net.ParseIP("192.168.1.2"), Port: 5678}
	data := []byte{10, 20, 30, 40, 255}

	result := script.OnData(from, to, data, nil)
	expected := []byte{11, 21, 31, 41, 0} // 255 + 1 = 256 % 256 = 0

	if result == nil {
		t.Fatal("Expected non-nil result when callback returns array of integers")
	}

	if len(result) != len(expected) {
		t.Fatalf("Expected result length %d, got %d", len(expected), len(result))
	}

	for i, b := range result {
		if b != expected[i] {
			t.Errorf("Expected byte at index %d to be %d, got %d", i, expected[i], b)
		}
	}
}
