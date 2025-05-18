package packets

import (
	"net"
	"reflect"
	"testing"
)

func TestQueueActivity(t *testing.T) {
	i := net.IP{}
	h := net.HardwareAddr{}
	a := Activity{
		IP:  i,
		MAC: h,
	}
	var units = []struct {
		got any
		exp any
	}{
		{a.IP, i},
		{a.MAC, h},
		{a.Source, false},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestQueueTraffic(t *testing.T) {
	tr := Traffic{}
	var units = []struct {
		got any
		exp any
	}{
		{tr.Sent, uint64(0)},
		{tr.Received, uint64(0)},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestQueueStats(t *testing.T) {
	s := Stats{}
	var units = []struct {
		got any
		exp any
	}{
		{s.Sent, uint64(0)},
		{s.Received, uint64(0)},
		{s.PktReceived, uint64(0)},
		{s.Errors, uint64(0)},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

// TODO: add tests for the rest of queue.go
