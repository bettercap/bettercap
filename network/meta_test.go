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

func TestMetaMarshalJSON(t *testing.T) {
	_, err := buildExampleMeta().MarshalJSON()
	if err != nil {
		t.Error("unable to marshal JSON from meta struct")
	}
}

func TestMetaSet(t *testing.T) {
	example := buildExampleMeta()
	example.Set("picat", "<3")
	if example.m["picat"] != "<3" {
		t.Error("unable to set meta data in struct")
	}
}
