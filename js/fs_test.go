package js

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/robertkrimen/otto"
)

func TestReadDir(t *testing.T) {
	vm := otto.New()

	// Create a temporary directory for testing
	tmpDir, err := ioutil.TempDir("", "js_test_readdir_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some test files and subdirectories
	testFiles := []string{"file1.txt", "file2.log", ".hidden"}
	testDirs := []string{"subdir1", "subdir2"}

	for _, name := range testFiles {
		if err := ioutil.WriteFile(filepath.Join(tmpDir, name), []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", name, err)
		}
	}

	for _, name := range testDirs {
		if err := os.Mkdir(filepath.Join(tmpDir, name), 0755); err != nil {
			t.Fatalf("failed to create test dir %s: %v", name, err)
		}
	}

	t.Run("valid directory", func(t *testing.T) {
		arg, _ := vm.ToValue(tmpDir)
		call := otto.FunctionCall{
			Otto:         vm,
			ArgumentList: []otto.Value{arg},
		}

		result := readDir(call)

		// Check if result is not undefined
		if result.IsUndefined() {
			t.Fatal("readDir returned undefined")
		}

		// Convert to Go slice
		export, err := result.Export()
		if err != nil {
			t.Fatalf("failed to export result: %v", err)
		}

		entries, ok := export.([]string)
		if !ok {
			t.Fatalf("expected []string, got %T", export)
		}

		// Check all expected entries are present
		expectedEntries := append(testFiles, testDirs...)
		if len(entries) != len(expectedEntries) {
			t.Errorf("expected %d entries, got %d", len(expectedEntries), len(entries))
		}

		// Check each entry exists
		for _, expected := range expectedEntries {
			found := false
			for _, entry := range entries {
				if entry == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected entry %s not found", expected)
			}
		}
	})

	t.Run("non-existent directory", func(t *testing.T) {
		arg, _ := vm.ToValue("/path/that/does/not/exist")
		call := otto.FunctionCall{
			Otto:         vm,
			ArgumentList: []otto.Value{arg},
		}

		result := readDir(call)

		// Should return undefined (error)
		if !result.IsUndefined() {
			t.Error("expected undefined for non-existent directory")
		}
	})

	t.Run("file instead of directory", func(t *testing.T) {
		// Create a file
		testFile := filepath.Join(tmpDir, "notadir.txt")
		ioutil.WriteFile(testFile, []byte("test"), 0644)

		arg, _ := vm.ToValue(testFile)
		call := otto.FunctionCall{
			Otto:         vm,
			ArgumentList: []otto.Value{arg},
		}

		result := readDir(call)

		// Should return undefined (error)
		if !result.IsUndefined() {
			t.Error("expected undefined when passing file instead of directory")
		}
	})

	t.Run("invalid arguments", func(t *testing.T) {
		tests := []struct {
			name string
			args []otto.Value
		}{
			{
				name: "no arguments",
				args: []otto.Value{},
			},
			{
				name: "too many arguments",
				args: func() []otto.Value {
					arg1, _ := vm.ToValue(tmpDir)
					arg2, _ := vm.ToValue("extra")
					return []otto.Value{arg1, arg2}
				}(),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				call := otto.FunctionCall{
					Otto:         vm,
					ArgumentList: tt.args,
				}

				result := readDir(call)

				// Should return undefined (error)
				if !result.IsUndefined() {
					t.Error("expected undefined for invalid arguments")
				}
			})
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		emptyDir := filepath.Join(tmpDir, "empty")
		os.Mkdir(emptyDir, 0755)

		arg, _ := vm.ToValue(emptyDir)
		call := otto.FunctionCall{
			Otto:         vm,
			ArgumentList: []otto.Value{arg},
		}

		result := readDir(call)

		if result.IsUndefined() {
			t.Fatal("readDir returned undefined for empty directory")
		}

		export, _ := result.Export()
		entries, _ := export.([]string)

		if len(entries) != 0 {
			t.Errorf("expected 0 entries for empty directory, got %d", len(entries))
		}
	})
}

func TestReadFile(t *testing.T) {
	vm := otto.New()

	// Create a temporary directory for testing
	tmpDir, err := ioutil.TempDir("", "js_test_readfile_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("valid file", func(t *testing.T) {
		testContent := "Hello, World!\nThis is a test file.\n特殊字符测试 🌍"
		testFile := filepath.Join(tmpDir, "test.txt")
		ioutil.WriteFile(testFile, []byte(testContent), 0644)

		arg, _ := vm.ToValue(testFile)
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{arg},
		}

		result := readFile(call)

		if result.IsUndefined() {
			t.Fatal("readFile returned undefined")
		}

		content, err := result.ToString()
		if err != nil {
			t.Fatalf("failed to convert result to string: %v", err)
		}

		if content != testContent {
			t.Errorf("expected content %q, got %q", testContent, content)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		arg, _ := vm.ToValue("/path/that/does/not/exist.txt")
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{arg},
		}

		result := readFile(call)

		// Should return undefined (error)
		if !result.IsUndefined() {
			t.Error("expected undefined for non-existent file")
		}
	})

	t.Run("directory instead of file", func(t *testing.T) {
		arg, _ := vm.ToValue(tmpDir)
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{arg},
		}

		result := readFile(call)

		// Should return undefined (error)
		if !result.IsUndefined() {
			t.Error("expected undefined when passing directory instead of file")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		emptyFile := filepath.Join(tmpDir, "empty.txt")
		ioutil.WriteFile(emptyFile, []byte(""), 0644)

		arg, _ := vm.ToValue(emptyFile)
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{arg},
		}

		result := readFile(call)

		if result.IsUndefined() {
			t.Fatal("readFile returned undefined for empty file")
		}

		content, _ := result.ToString()
		if content != "" {
			t.Errorf("expected empty string, got %q", content)
		}
	})

	t.Run("binary file", func(t *testing.T) {
		binaryContent := []byte{0, 1, 2, 3, 255, 254, 253, 252}
		binaryFile := filepath.Join(tmpDir, "binary.bin")
		ioutil.WriteFile(binaryFile, binaryContent, 0644)

		arg, _ := vm.ToValue(binaryFile)
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{arg},
		}

		result := readFile(call)

		if result.IsUndefined() {
			t.Fatal("readFile returned undefined for binary file")
		}

		content, _ := result.ToString()
		if content != string(binaryContent) {
			t.Error("binary content mismatch")
		}
	})

	t.Run("invalid arguments", func(t *testing.T) {
		tests := []struct {
			name string
			args []otto.Value
		}{
			{
				name: "no arguments",
				args: []otto.Value{},
			},
			{
				name: "too many arguments",
				args: func() []otto.Value {
					arg1, _ := vm.ToValue("file.txt")
					arg2, _ := vm.ToValue("extra")
					return []otto.Value{arg1, arg2}
				}(),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				call := otto.FunctionCall{
					ArgumentList: tt.args,
				}

				result := readFile(call)

				// Should return undefined (error)
				if !result.IsUndefined() {
					t.Error("expected undefined for invalid arguments")
				}
			})
		}
	})

	t.Run("large file", func(t *testing.T) {
		// Create a 1MB file
		largeContent := strings.Repeat("A", 1024*1024)
		largeFile := filepath.Join(tmpDir, "large.txt")
		ioutil.WriteFile(largeFile, []byte(largeContent), 0644)

		arg, _ := vm.ToValue(largeFile)
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{arg},
		}

		result := readFile(call)

		if result.IsUndefined() {
			t.Fatal("readFile returned undefined for large file")
		}

		content, _ := result.ToString()
		if len(content) != len(largeContent) {
			t.Errorf("expected content length %d, got %d", len(largeContent), len(content))
		}
	})
}

func TestWriteFile(t *testing.T) {
	vm := otto.New()

	// Create a temporary directory for testing
	tmpDir, err := ioutil.TempDir("", "js_test_writefile_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("write new file", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "new_file.txt")
		testContent := "Hello, World!\nThis is a new file.\n特殊字符测试 🌍"

		argFile, _ := vm.ToValue(testFile)
		argContent, _ := vm.ToValue(testContent)
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{argFile, argContent},
		}

		result := writeFile(call)

		// writeFile returns null on success
		if !result.IsNull() {
			t.Error("expected null return value for successful write")
		}

		// Verify file was created with correct content
		content, err := ioutil.ReadFile(testFile)
		if err != nil {
			t.Fatalf("failed to read written file: %v", err)
		}

		if string(content) != testContent {
			t.Errorf("expected content %q, got %q", testContent, string(content))
		}

		// Check file permissions
		info, _ := os.Stat(testFile)
		if runtime.GOOS == "windows" {
			// On Windows, permissions are different - just check that file exists and is readable
			if info.Mode()&0400 == 0 {
				t.Error("expected file to be readable on Windows")
			}
		} else {
			// On Unix-like systems, check exact permissions
			if info.Mode().Perm() != 0644 {
				t.Errorf("expected permissions 0644, got %v", info.Mode().Perm())
			}
		}
	})

	t.Run("overwrite existing file", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "existing.txt")
		oldContent := "Old content"
		newContent := "New content that is longer than the old content"

		// Create initial file
		ioutil.WriteFile(testFile, []byte(oldContent), 0644)

		argFile, _ := vm.ToValue(testFile)
		argContent, _ := vm.ToValue(newContent)
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{argFile, argContent},
		}

		result := writeFile(call)

		if !result.IsNull() {
			t.Error("expected null return value for successful write")
		}

		// Verify file was overwritten
		content, _ := ioutil.ReadFile(testFile)
		if string(content) != newContent {
			t.Errorf("expected content %q, got %q", newContent, string(content))
		}
	})

	t.Run("write to non-existent directory", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "nonexistent", "subdir", "file.txt")
		testContent := "test"

		argFile, _ := vm.ToValue(testFile)
		argContent, _ := vm.ToValue(testContent)
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{argFile, argContent},
		}

		result := writeFile(call)

		// Should return undefined (error)
		if !result.IsUndefined() {
			t.Error("expected undefined when writing to non-existent directory")
		}
	})

	t.Run("write empty content", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "empty.txt")

		argFile, _ := vm.ToValue(testFile)
		argContent, _ := vm.ToValue("")
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{argFile, argContent},
		}

		result := writeFile(call)

		if !result.IsNull() {
			t.Error("expected null return value for successful write")
		}

		// Verify empty file was created
		content, _ := ioutil.ReadFile(testFile)
		if len(content) != 0 {
			t.Errorf("expected empty file, got %d bytes", len(content))
		}
	})

	t.Run("invalid arguments", func(t *testing.T) {
		tests := []struct {
			name string
			args []otto.Value
		}{
			{
				name: "no arguments",
				args: []otto.Value{},
			},
			{
				name: "one argument",
				args: func() []otto.Value {
					arg, _ := vm.ToValue("file.txt")
					return []otto.Value{arg}
				}(),
			},
			{
				name: "too many arguments",
				args: func() []otto.Value {
					arg1, _ := vm.ToValue("file.txt")
					arg2, _ := vm.ToValue("content")
					arg3, _ := vm.ToValue("extra")
					return []otto.Value{arg1, arg2, arg3}
				}(),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				call := otto.FunctionCall{
					ArgumentList: tt.args,
				}

				result := writeFile(call)

				// Should return undefined (error)
				if !result.IsUndefined() {
					t.Error("expected undefined for invalid arguments")
				}
			})
		}
	})

	t.Run("write binary content", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "binary.bin")
		binaryContent := string([]byte{0, 1, 2, 3, 255, 254, 253, 252})

		argFile, _ := vm.ToValue(testFile)
		argContent, _ := vm.ToValue(binaryContent)
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{argFile, argContent},
		}

		result := writeFile(call)

		if !result.IsNull() {
			t.Error("expected null return value for successful write")
		}

		// Verify binary content
		content, _ := ioutil.ReadFile(testFile)
		if string(content) != binaryContent {
			t.Error("binary content mismatch")
		}
	})
}

func TestFileSystemIntegration(t *testing.T) {
	vm := otto.New()

	// Create a temporary directory for testing
	tmpDir, err := ioutil.TempDir("", "js_test_integration_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("write then read file", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "roundtrip.txt")
		testContent := "Round-trip test content\nLine 2\nLine 3"

		// Write file
		argFile, _ := vm.ToValue(testFile)
		argContent, _ := vm.ToValue(testContent)
		writeCall := otto.FunctionCall{
			ArgumentList: []otto.Value{argFile, argContent},
		}

		writeResult := writeFile(writeCall)
		if !writeResult.IsNull() {
			t.Fatal("write failed")
		}

		// Read file back
		readCall := otto.FunctionCall{
			ArgumentList: []otto.Value{argFile},
		}

		readResult := readFile(readCall)
		if readResult.IsUndefined() {
			t.Fatal("read failed")
		}

		readContent, _ := readResult.ToString()
		if readContent != testContent {
			t.Errorf("round-trip failed: expected %q, got %q", testContent, readContent)
		}
	})

	t.Run("create files then list directory", func(t *testing.T) {
		// Create multiple files
		files := []string{"file1.txt", "file2.txt", "file3.txt"}
		for _, name := range files {
			path := filepath.Join(tmpDir, name)
			argFile, _ := vm.ToValue(path)
			argContent, _ := vm.ToValue("content of " + name)
			call := otto.FunctionCall{
				ArgumentList: []otto.Value{argFile, argContent},
			}
			writeFile(call)
		}

		// List directory
		argDir, _ := vm.ToValue(tmpDir)
		listCall := otto.FunctionCall{
			Otto:         vm,
			ArgumentList: []otto.Value{argDir},
		}

		listResult := readDir(listCall)
		if listResult.IsUndefined() {
			t.Fatal("readDir failed")
		}

		export, _ := listResult.Export()
		entries, _ := export.([]string)

		// Check all files are listed
		for _, expected := range files {
			found := false
			for _, entry := range entries {
				if entry == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected file %s not found in directory listing", expected)
			}
		}
	})
}

func BenchmarkReadFile(b *testing.B) {
	vm := otto.New()

	// Create test file
	tmpFile, _ := ioutil.TempFile("", "bench_readfile_*")
	defer os.Remove(tmpFile.Name())

	content := strings.Repeat("Benchmark test content line\n", 100)
	ioutil.WriteFile(tmpFile.Name(), []byte(content), 0644)

	arg, _ := vm.ToValue(tmpFile.Name())
	call := otto.FunctionCall{
		ArgumentList: []otto.Value{arg},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = readFile(call)
	}
}

func BenchmarkWriteFile(b *testing.B) {
	vm := otto.New()

	tmpDir, _ := ioutil.TempDir("", "bench_writefile_*")
	defer os.RemoveAll(tmpDir)

	content := strings.Repeat("Benchmark test content line\n", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("bench_%d.txt", i))
		argFile, _ := vm.ToValue(testFile)
		argContent, _ := vm.ToValue(content)
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{argFile, argContent},
		}
		_ = writeFile(call)
	}
}

func BenchmarkReadDir(b *testing.B) {
	vm := otto.New()

	// Create test directory with files
	tmpDir, _ := ioutil.TempDir("", "bench_readdir_*")
	defer os.RemoveAll(tmpDir)

	// Create 100 files
	for i := 0; i < 100; i++ {
		name := filepath.Join(tmpDir, fmt.Sprintf("file_%d.txt", i))
		ioutil.WriteFile(name, []byte("test"), 0644)
	}

	arg, _ := vm.ToValue(tmpDir)
	call := otto.FunctionCall{
		Otto:         vm,
		ArgumentList: []otto.Value{arg},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = readDir(call)
	}
}
