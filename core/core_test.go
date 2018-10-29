package core

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/str"
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

func TestCoreExec(t *testing.T) {
	var units = []struct {
		exec   string
		args   []string
		out    string
		err    string
		stdout string
	}{
		{"foo", []string{}, "", `exec: "foo": executable file not found in $PATH`, `ERROR for 'foo []': exec: "foo": executable file not found in $PATH`},
		{"ps", []string{"-someinvalidflag"}, "", "exit status 1", "ERROR for 'ps [-someinvalidflag]': exit status 1"},
		{"true", []string{}, "", "", ""},
		{"head", []string{"/path/to/file/that/does/not/exist"}, "", "exit status 1", "ERROR for 'head [/path/to/file/that/does/not/exist]': exit status 1"},
	}

	for _, u := range units {
		var buf bytes.Buffer

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		gotOut, gotErr := ExecSilent(u.exec, u.args)
		w.Close()
		io.Copy(&buf, r)
		os.Stdout = oldStdout

		gotStdout := str.Trim(buf.String())
		if gotOut != u.out {
			t.Fatalf("expected output '%s', got '%s'", u.out, gotOut)
		} else if u.err == "" && gotErr != nil {
			t.Fatalf("expected no error, got '%s'", gotErr)
		} else if u.err != "" && gotErr == nil {
			t.Fatalf("expected error '%s', got none", u.err)
		} else if u.err != "" && gotErr != nil && gotErr.Error() != u.err {
			t.Fatalf("expected error '%s', got '%s'", u.err, gotErr)
		} else if gotStdout != "" {
			t.Fatalf("expected empty stdout, got '%s'", gotStdout)
		}
	}

	for _, u := range units {
		var buf bytes.Buffer

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		gotOut, gotErr := Exec(u.exec, u.args)
		w.Close()
		io.Copy(&buf, r)
		os.Stdout = oldStdout

		gotStdout := str.Trim(buf.String())
		if gotOut != u.out {
			t.Fatalf("expected output '%s', got '%s'", u.out, gotOut)
		} else if u.err == "" && gotErr != nil {
			t.Fatalf("expected no error, got '%s'", gotErr)
		} else if u.err != "" && gotErr == nil {
			t.Fatalf("expected error '%s', got none", u.err)
		} else if u.err != "" && gotErr != nil && gotErr.Error() != u.err {
			t.Fatalf("expected error '%s', got '%s'", u.err, gotErr)
		} else if gotStdout != u.stdout {
			t.Fatalf("expected stdout '%s', got '%s'", u.stdout, gotStdout)
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
