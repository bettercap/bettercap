package caplets

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestNewCaplet(t *testing.T) {
	name := "test-caplet"
	path := "/path/to/caplet.cap"
	size := int64(1024)

	cap := NewCaplet(name, path, size)

	if cap.Name != name {
		t.Errorf("expected name %s, got %s", name, cap.Name)
	}
	if cap.Path != path {
		t.Errorf("expected path %s, got %s", path, cap.Path)
	}
	if cap.Size != size {
		t.Errorf("expected size %d, got %d", size, cap.Size)
	}
	if cap.Code == nil {
		t.Error("Code should not be nil")
	}
	if cap.Scripts == nil {
		t.Error("Scripts should not be nil")
	}
}

func TestCapletEval(t *testing.T) {
	tests := []struct {
		name      string
		code      []string
		argv      []string
		wantLines []string
		wantErr   bool
	}{
		{
			name:      "empty code",
			code:      []string{},
			argv:      nil,
			wantLines: []string{},
			wantErr:   false,
		},
		{
			name: "skip comments and empty lines",
			code: []string{
				"# this is a comment",
				"",
				"set test value",
				"# another comment",
				"set another value",
			},
			argv: nil,
			wantLines: []string{
				"set test value",
				"set another value",
			},
			wantErr: false,
		},
		{
			name: "variable substitution",
			code: []string{
				"set param $0",
				"set value $1",
				"run $0 $1 $2",
			},
			argv: []string{"arg0", "arg1", "arg2"},
			wantLines: []string{
				"set param arg0",
				"set value arg1",
				"run arg0 arg1 arg2",
			},
			wantErr: false,
		},
		{
			name: "multiple occurrences of same variable",
			code: []string{
				"$0 $0 $1 $0",
			},
			argv: []string{"foo", "bar"},
			wantLines: []string{
				"foo foo bar foo",
			},
			wantErr: false,
		},
		{
			name: "missing argv values",
			code: []string{
				"set $0 $1 $2",
			},
			argv: []string{"only_one"},
			wantLines: []string{
				"set only_one $1 $2",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFile, err := ioutil.TempFile("", "test-*.cap")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tempFile.Name())
			tempFile.Close()

			cap := NewCaplet("test", tempFile.Name(), 100)
			cap.Code = tt.code

			var gotLines []string
			err = cap.Eval(tt.argv, func(line string) error {
				gotLines = append(gotLines, line)
				return nil
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("Eval() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(gotLines) != len(tt.wantLines) {
				t.Errorf("got %d lines, want %d", len(gotLines), len(tt.wantLines))
				return
			}

			for i, line := range gotLines {
				if line != tt.wantLines[i] {
					t.Errorf("line %d: got %q, want %q", i, line, tt.wantLines[i])
				}
			}
		})
	}
}

func TestCapletEvalError(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "test-*.cap")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	cap := NewCaplet("test", tempFile.Name(), 100)
	cap.Code = []string{
		"first line",
		"error line",
		"third line",
	}

	expectedErr := errors.New("test error")
	var executedLines []string

	err = cap.Eval(nil, func(line string) error {
		executedLines = append(executedLines, line)
		if line == "error line" {
			return expectedErr
		}
		return nil
	})

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	// Should have executed first two lines before error
	if len(executedLines) != 2 {
		t.Errorf("expected 2 executed lines, got %d", len(executedLines))
	}
}

func TestCapletEvalWithChdirPath(t *testing.T) {
	// Create a temporary caplet file to test with
	tempFile, err := ioutil.TempFile("", "test-*.cap")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	cap := NewCaplet("test", tempFile.Name(), 100)
	cap.Code = []string{"test command"}

	executed := false
	err = cap.Eval(nil, func(line string) error {
		executed = true
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !executed {
		t.Error("callback was not executed")
	}
}

func TestNewScript(t *testing.T) {
	path := "/path/to/script.js"
	size := int64(2048)

	script := newScript(path, size)

	if script.Path != path {
		t.Errorf("expected path %s, got %s", path, script.Path)
	}
	if script.Size != size {
		t.Errorf("expected size %d, got %d", size, script.Size)
	}
	if script.Code == nil {
		t.Error("Code should not be nil")
	}
	if len(script.Code) != 0 {
		t.Errorf("expected empty Code slice, got %d elements", len(script.Code))
	}
}

func TestCapletEvalCommentAtStartOfLine(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "test-*.cap")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	cap := NewCaplet("test", tempFile.Name(), 100)
	cap.Code = []string{
		"# comment",
		" # not a comment (has space before #)",
		"	# not a comment (has tab before #)",
		"command # inline comment",
	}

	var gotLines []string
	err = cap.Eval(nil, func(line string) error {
		gotLines = append(gotLines, line)
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedLines := []string{
		" # not a comment (has space before #)",
		"	# not a comment (has tab before #)",
		"command # inline comment",
	}

	if len(gotLines) != len(expectedLines) {
		t.Errorf("got %d lines, want %d", len(gotLines), len(expectedLines))
		return
	}

	for i, line := range gotLines {
		if line != expectedLines[i] {
			t.Errorf("line %d: got %q, want %q", i, line, expectedLines[i])
		}
	}
}

func TestCapletEvalArgvSubstitutionEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		argv     []string
		wantLine string
	}{
		{
			name:     "double digit substitution $10",
			code:     "$1$0",
			argv:     []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
			wantLine: "10",
		},
		{
			name:     "no space between variables",
			code:     "$0$1$2",
			argv:     []string{"a", "b", "c"},
			wantLine: "abc",
		},
		{
			name:     "variables in quotes",
			code:     `"$0" '$1'`,
			argv:     []string{"foo", "bar"},
			wantLine: `"foo" 'bar'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFile, err := ioutil.TempFile("", "test-*.cap")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tempFile.Name())
			tempFile.Close()

			cap := NewCaplet("test", tempFile.Name(), 100)
			cap.Code = []string{tt.code}

			var gotLine string
			err = cap.Eval(tt.argv, func(line string) error {
				gotLine = line
				return nil
			})

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if gotLine != tt.wantLine {
				t.Errorf("got line %q, want %q", gotLine, tt.wantLine)
			}
		})
	}
}

func TestCapletStructFields(t *testing.T) {
	// Test that Caplet properly embeds Script
	tempFile, err := ioutil.TempFile("", "test-*.cap")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	cap := NewCaplet("test", tempFile.Name(), 100)

	// These fields should be accessible due to embedding
	_ = cap.Path
	_ = cap.Size
	_ = cap.Code

	// And these are Caplet's own fields
	_ = cap.Name
	_ = cap.Scripts
}

func BenchmarkCapletEval(b *testing.B) {
	cap := NewCaplet("bench", "/tmp/bench.cap", 100)
	cap.Code = []string{
		"set param1 $0",
		"set param2 $1",
		"# comment line",
		"",
		"run command $0 $1 $2",
		"another command",
	}
	argv := []string{"arg0", "arg1", "arg2"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cap.Eval(argv, func(line string) error {
			// Do nothing, just measure evaluation overhead
			return nil
		})
	}
}

func BenchmarkVariableSubstitution(b *testing.B) {
	line := "command $0 $1 $2 $3 $4 $5 $6 $7 $8 $9"
	argv := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := line
		for j, arg := range argv {
			what := "$" + string(rune('0'+j))
			result = strings.Replace(result, what, arg, -1)
		}
	}
}
