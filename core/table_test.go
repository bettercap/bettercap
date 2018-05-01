package core

import (
	"bytes"
	"regexp"
	"testing"
)

func TestViewLen(t *testing.T) {
	exp := 2
	got := viewLen("<3")
	if got != exp {
		t.Fatalf("expected '%d', got '%d'", exp, got)
	}
}

func TestMaxLen(t *testing.T) {
	exp := 7
	got := maxLen([]string{"go", "python", "ruby", "crystal"})
	if got != exp {
		t.Fatalf("expected '%d', got '%d'", exp, got)
	}
}

func TestAlignLeft(t *testing.T) {
	exp := Alignment(0)
	got := AlignLeft
	if got != exp {
		t.Fatalf("expected '%d', got '%d'", exp, got)
	}
}

func TestAlignCenter(t *testing.T) {
	exp := Alignment(1)
	got := AlignCenter
	if got != exp {
		t.Fatalf("expected '%d', got '%d'", exp, got)
	}
}

func TestAlignRight(t *testing.T) {
	exp := Alignment(2)
	got := AlignRight
	if got != exp {
		t.Fatalf("expected '%d', got '%d'", exp, got)
	}
}

func TestGetPads(t *testing.T) {
	lPadExp := -9
	rPadExp := -8
	lPadGot, rPadGot := getPads("Pikachu, thunderbolt!", 3, AlignCenter)
	if rPadGot != rPadExp {
		t.Fatalf("expected '%d', got '%d'", rPadExp, rPadGot)
	}
	if lPadGot != lPadExp {
		t.Fatalf("expected '%d', got '%d'", lPadExp, lPadGot)
	}
}

func TestPadded(t *testing.T) {
	exp := "<3"
	got := padded("<3", 1, AlignLeft)
	if got != exp {
		t.Fatalf("expected '%s', got '%s'", exp, got)
	}
}

func TestAsTable(t *testing.T) {
	var b bytes.Buffer

	AsTable(&b, []string{"top"}, [][]string{[]string{"bottom"}})

	// look for "+-------+" table style for whichever size
	match, err := regexp.MatchString(`\++-+\+`, b.String())
	if err != nil {
		t.Fatalf("unable to perform regex on table output")
	}
	if !match {
		t.Fatalf("expected table in format not found")
	}
}
