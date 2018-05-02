package network

import "testing"

func buildExampleAliases() *Aliases {
	return &Aliases{}
}

func TestAliasesLoadAliases(t *testing.T) {
	err, _ := LoadAliases()
	if err != nil {
		t.Error(err)
	}
}

func TestAliasesSaveUnlocked(t *testing.T) {
	exampleAliases := buildExampleAliases()
	err := exampleAliases.saveUnlocked()
	if err != nil {
		t.Error(err)
	}
}

func TestAliasesSave(t *testing.T) {
	exampleAliases := buildExampleAliases()
	err := exampleAliases.Save()
	if err != nil {
		t.Error(err)
	}
}

func TestAliasesGet(t *testing.T) {
	exampleAliases := buildExampleAliases()

	exp := ""
	got := exampleAliases.Get("pi:ca:tw:as:he:re")

	if got != exp {
		t.Fatalf("expected '%v', got '%v'", exp, got)
	}
}

func TestAliasesSet(t *testing.T) {
	exampleAliases := buildExampleAliases()
	exampleAliases.data = make(map[string]string)

	if exampleAliases.Set("pi:ca:tw:as:he:re", "picat") != nil {
		t.Error("unable to set alias")
	}

	if exampleAliases.Get("pi:ca:tw:as:he:re") != "picat" {
		t.Error("unable to get set alias")
	}
}

func TestAliasesFind(t *testing.T) {
	exampleAliases := buildExampleAliases()
	exampleAliases.data = make(map[string]string)
	err := exampleAliases.Set("pi:ca:tw:as:he:re", "picat")
	if err != nil {
		t.Error(err)
	}
	mac, found := exampleAliases.Find("picat")
	if !found {
		t.Error("unable to find mac address for alias")
	}
	if mac != "pi:ca:tw:as:he:re" {
		t.Error("unable to find correct mac address for alias")
	}
}
