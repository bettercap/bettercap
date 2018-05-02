package network

import (
	"strings"
	"testing"
)

func buildExampleMeta() *Meta {
	return NewMeta()
}

func TestNewMeta(t *testing.T) {
	exp := len(Meta{}.m)
	got := len(NewMeta().m)
	if got != exp {
		t.Fatalf("expected '%v', got '%v'", exp, got)
	}
}
