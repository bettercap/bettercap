package core

import (
	"os"
	"testing"

	"github.com/evilsocket/islazy/fs"
)

func hasInt(a []int, v int) bool {
	for _, n := range a {
		if n == v {
			return true
		}
	}
	return false
}

func sameInts(a []int, b []int, ordered bool) bool {
	if len(a) != len(b) {
		return false
	}

	if ordered {
		for i, v := range a {
			if v != b[i] {
				return false
			}
		}
	} else {
		for _, v := range a {
			if !hasInt(b, v) {
				return false
			}
		}
	}

	return true
}

func TestCoreUniqueIntsUnsorted(t *testing.T) {
	var units = []struct {
		from []int
		to   []int
	}{
		{[]int{}, []int{}},
		{[]int{1, 1, 1, 1, 1}, []int{1}},
		{[]int{1, 2, 1, 2, 3, 4}, []int{1, 2, 3, 4}},
		{[]int{4, 3, 4, 3, 2, 2}, []int{4, 3, 2}},
		{[]int{8, 3, 8, 4, 6, 1}, []int{8, 3, 4, 6, 1}},
	}

	for _, u := range units {
		got := UniqueInts(u.from, false)
		if !sameInts(got, u.to, false) {
			t.Fatalf("expected '%v', got '%v'", u.to, got)
		}
	}
}

func TestCoreUniqueIntsSorted(t *testing.T) {
	var units = []struct {
		from []int
		to   []int
	}{
		{[]int{}, []int{}},
		{[]int{1, 1, 1, 1, 1}, []int{1}},
		{[]int{1, 2, 1, 2, 3, 4}, []int{1, 2, 3, 4}},
		{[]int{4, 3, 4, 3, 2, 2}, []int{2, 3, 4}},
		{[]int{8, 3, 8, 4, 6, 1}, []int{1, 3, 4, 6, 8}},
	}

	for _, u := range units {
		got := UniqueInts(u.from, true)
		if !sameInts(got, u.to, true) {
			t.Fatalf("expected '%v', got '%v'", u.to, got)
		}
	}
}

func TestCoreExists(t *testing.T) {
	var units = []struct {
		what   string
		exists bool
	}{
		{".", true},
		{"/", true},
		{"wuuut", false},
		{"/wuuu.t", false},
		{os.Args[0], true},
	}

	for _, u := range units {
		got := fs.Exists(u.what)
		if got != u.exists {
			t.Fatalf("expected '%v', got '%v'", u.exists, got)
		}
	}
}
