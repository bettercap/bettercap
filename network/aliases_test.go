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

func TestAliasesSaveUnlocked(t *testing.T) {
	exampleAliases := buildExampleAlaises()
	err := exampleAliases.saveUnlocked()
	if err != nil {
		t.Error(err)
	}
}
