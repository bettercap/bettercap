package core

import "testing"

func TestW(t *testing.T) {
	exp := "<3\033[0m"
	got := W("<", "3")
	if got != exp {
		t.Fatalf("expected '%s', got '%s'", exp, got)
	}
}

func TestBold(t *testing.T) {
	exp := "\033[1mgohpers\033[0m"
	got := Bold("gohpers")
	if got != exp {
		t.Fatalf("expected path '%s', got '%s'", exp, got)
	}
}

func TestDim(t *testing.T) {
	exp := "\033[2mgohpers\033[0m"
	got := Dim("gohpers")
	if got != exp {
		t.Fatalf("expected path '%s', got '%s'", exp, got)
	}
}

func TestRed(t *testing.T) {
	exp := "\033[31mgohpers\033[0m"
	got := Red("gohpers")
	if got != exp {
		t.Fatalf("expected path '%s', got '%s'", exp, got)
	}
}

func TestGreen(t *testing.T) {
	exp := "\033[32mgohpers\033[0m"
	got := Green("gohpers")
	if got != exp {
		t.Fatalf("expected path '%s', got '%s'", exp, got)
	}
}

func TestBlue(t *testing.T) {
	exp := "\033[34mgohpers\033[0m"
	got := Blue("gohpers")
	if got != exp {
		t.Fatalf("expected path '%s', got '%s'", exp, got)
	}
}

func TestYellow(t *testing.T) {
	exp := "\033[33mgohpers\033[0m"
	got := Yellow("gohpers")
	if got != exp {
		t.Fatalf("expected path '%s', got '%s'", exp, got)
	}
}
