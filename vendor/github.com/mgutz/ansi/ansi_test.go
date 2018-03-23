package ansi

import (
	"strings"
	"testing"
)

func TestPlain(t *testing.T) {
	DisableColors(true)
	PrintStyles()
}

func TestStyles(t *testing.T) {
	DisableColors(false)
	PrintStyles()
}

func TestDisableColors(t *testing.T) {
	fn := ColorFunc("red")

	buf := colorCode("off")
	if buf.String() != "" {
		t.Fail()
	}

	DisableColors(true)
	if Black != "" {
		t.Fail()
	}
	code := ColorCode("red")
	if code != "" {
		t.Fail()
	}
	s := fn("foo")
	if s != "foo" {
		t.Fail()
	}

	DisableColors(false)
	if Black == "" {
		t.Fail()
	}
	code = ColorCode("red")
	if code == "" {
		t.Fail()
	}
	// will have escape codes around it
	index := strings.Index(fn("foo"), "foo")
	if index <= 0 {
		t.Fail()
	}
}
