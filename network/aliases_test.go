package network

import "testing"

func buildExampleAlaises() *Aliases {
	return &Aliases{}
}

func TestAliasesLoadAliases(t *testing.T) {
	err, _ := LoadAliases()
	if err != nil {
		t.Error(err)
	}
}
