package session

import (
	"testing"

	"github.com/bettercap/bettercap/v2/core"
)

func TestParamGettersMissingParameter(t *testing.T) {
	m := NewSessionModule("test", &Session{})

	if err, _ := m.StringParam("missing"); err == nil {
		t.Error("StringParam missing: expected error")
	}
	if err, _ := m.IntParam("missing"); err == nil {
		t.Error("IntParam missing: expected error")
	}
	if err, _ := m.DecParam("missing"); err == nil {
		t.Error("DecParam missing: expected error")
	}
	if err, _ := m.BoolParam("missing"); err == nil {
		t.Error("BoolParam missing: expected error")
	}
}

func TestBoolParamExistingParameter(t *testing.T) {
	s, err := WithOptions(core.Options{})
	if err != nil {
		t.Fatalf("WithOptions: %v", err)
	}
	defer s.Close()

	m := NewSessionModule("test", s)
	m.AddParam(NewBoolParameter("test.bool", "true", "test"))

	if err, value := m.BoolParam("test.bool"); err != nil {
		t.Fatalf("BoolParam existing: %v", err)
	} else if !value {
		t.Fatal("BoolParam existing: expected true")
	}
}
