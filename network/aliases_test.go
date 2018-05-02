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

func TestAliasesSave(t *testing.T) {
	exampleAliases := buildExampleAlaises()
	err := exampleAliases.Save()
	if err != nil {
		t.Error(err)
	}
}

func TestAliasesGet(t *testing.T) {
	exampleAliases := buildExampleAlaises()

	exp := ""
	got := exampleAliases.Get("pi:ca:tw:as:he:re")

	if got != exp {
		t.Fatalf("expected '%v', got '%v'", exp, got)
	}
}
