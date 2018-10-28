package session

import (
	"testing"
)

func sameStrings(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func assertPanic(t *testing.T, msg string, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal(msg)
		}
	}()
	f()
}

func TestSessionCommandHandler(t *testing.T) {
	var units = []struct {
		expr   string
		panic  bool
		parsed []string
	}{
		{"notvali(d", true, nil},
		{`simple\s+(\d+)`, false, []string{"123"}},
	}

	for _, u := range units {
		if u.panic {
			assertPanic(t, "", func() {
				_ = NewCommandHandler("", u.expr, "", nil)
				t.Fatal("panic expected")
			})
		} else {
			c := NewCommandHandler("", u.expr, "", nil)
			shouldNotParse := "simple123"
			shouldParse := "simple 123"

			if parsed, _ := c.Parse(shouldNotParse); parsed {
				t.Fatalf("should not parse '%s'", shouldNotParse)
			} else if parsed, parts := c.Parse(shouldParse); !parsed {
				t.Fatalf("should parse '%s'", shouldParse)
			} else if !sameStrings(parts, u.parsed) {
				t.Fatalf("expected '%v', got '%v'", u.parsed, parts)
			}
		}
	}
}
