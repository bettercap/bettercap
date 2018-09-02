package core

import (
	"os"
	"testing"
)

func TestIsDumbTerminal(t *testing.T) {
	term := os.Getenv("TERM")
	os.Setenv("TERM", "dumb")
	if !isDumbTerminal() {
		t.Fatal("Expected false when TERM==dumb")
	}
	os.Setenv("TERM", "")
	if !isDumbTerminal() {
		t.Fatal("Expected false when TERM empty")
	}
	os.Setenv("TERM", term)
}

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

func TestInitSwag(t *testing.T) {
	// Run after other tests to avoid breaking globals
	// Test InitSwag unsets globals when set
	BOLD = "\033[1m"
	InitSwag(true)
	if BOLD != "" {
		t.Fatal("expected BOLD to be empty string")
	}
	term := os.Getenv("TERM")
	os.Setenv("TERM", "dumb")
	BOLD = "\033[1m"
	InitSwag(false)
	if BOLD != "" {
		t.Fatal("expected BOLD to be empty string")
	}
	os.Setenv("TERM", term)
	// Would be good to test BOLD *isn't* unset when we have a TTY
	// but less trivial to stub os.File.Fd() without complicating architecture
}
