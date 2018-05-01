package network

import "testing"

func TestOuiVar(t *testing.T) {
	if len(oui) <= 0 {
		t.Error("unable to find any oui infromation")
	}
}
