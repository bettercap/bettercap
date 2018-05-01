package network

import "testing"

func TestOuiVar(t *testing.T) {
	if len(oui) <= 0 {
		t.Error("unable to find any oui infromation")
	}
}

func TestOuiLookup(t *testing.T) {
	exampleMac := "e0:0c:7f:XX:XX:XX"
	exp := "Nintendo Co."
	got := OuiLookup(exampleMac)
	if got != exp {
		t.Fatalf("expected '%s', got '%s'", exp, got)
	}
}
