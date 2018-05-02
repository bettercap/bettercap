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

func TestAliasesSet(t *testing.T) {
	exampleAliases := buildExampleAlaises()
	exampleAliases.data = make(map[string]string)

	if exampleAliases.Set("pi:ca:tw:as:he:re", "picat") != nil {
		t.Error("unable to set alias")
	}

	if exampleAliases.Get("pi:ca:tw:as:he:re") != "picat" {
		t.Error("unable to get set alias")
	}
}
