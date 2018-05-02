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

// TODO document what this does, not too clear,
// at least for me today lolololol
func TestMetaGetIntsWith(t *testing.T) {
	example := buildExampleMeta()
	example.m["picat"] = "3,"

	exp := []int{4, 3}
	got := example.GetIntsWith("picat", 4, false)

	if len(exp) != len(got) {
		t.Fatalf("expected '%v', got '%v'", exp, got)
	}
}

func TestMetaSetInts(t *testing.T) {
	example := buildExampleMeta()
	example.SetInts("picat", []int{0, 1})

	exp := strings.Join([]string{"0", "1"}, ",")
	got := example.m["picat"]

	if exp != got {
		t.Fatalf("expected '%v', got '%v'", exp, got)
	}
}

func TestMetaGetOr(t *testing.T) {
	example := buildExampleMeta()
	dflt := "picat"
	exp := dflt
	got := example.GetOr("evilsocket", dflt)
	if exp != got {
		t.Fatalf("expected '%v', got '%v'", exp, got)
	}
}

func TestMetaEach(t *testing.T) {
	example := buildExampleMeta()
	example.m["picat"] = true
	example.m["evilsocket"] = true

	count := 0
	exampleCB := func(name string, value interface{}) {
		count++
	}
	example.Each(exampleCB)

	exp := 2
	got := count
	if exp != got {
		t.Fatalf("expected '%v', got '%v'", exp, got)
	}
}

func TestMetaEmpty(t *testing.T) {
	example := buildExampleMeta()

	if !example.Empty() {
		t.Error("unable to check if filled struct is empty")
	}

	example.m["picat"] = true //fill struct so not empty

	if example.Empty() {
		t.Error("unable to check if filled struct is empty")
	}

}
