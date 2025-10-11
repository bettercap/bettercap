package js

import (
	"fmt"
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
	tmpDir, err := os.MkdirTemp("", "js_test_readdir_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some test files and subdirectories
	testFiles := []string{"file1.txt", "file2.log", ".hidden"}
	testDirs := []string{"subdir1", "subdir2"}

	for _, name := range testFiles {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("test"), 0644); err != nil {
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
		os.WriteFile(testFile, []byte("test"), 0644)

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
	tmpDir, err := os.MkdirTemp("", "js_test_readfile_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("valid file", func(t *testing.T) {
		testContent := "Hello, World!\nThis is a test file.\n特殊字符测试 🌍"
		testFile := filepath.Join(tmpDir, "test.txt")
		os.WriteFile(testFile, []byte(testContent), 0644)

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
		os.WriteFile(emptyFile, []byte(""), 0644)

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
		os.WriteFile(binaryFile, binaryContent, 0644)

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
		os.WriteFile(largeFile, []byte(largeContent), 0644)

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
	tmpDir, err := os.MkdirTemp("", "js_test_writefile_*")
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
		content, err := os.ReadFile(testFile)
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
		os.WriteFile(testFile, []byte(oldContent), 0644)

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
		content, _ := os.ReadFile(testFile)
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
		content, _ := os.ReadFile(testFile)
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
		content, _ := os.ReadFile(testFile)
		if string(content) != binaryContent {
			t.Error("binary content mismatch")
		}
	})
}

func TestAppendFile(t *testing.T) {
	vm := otto.New()

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "js_test_appendfile_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("append to new file", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "new_append.txt")
		testContent := "Hello, World!\nThis is appended content.\n特殊字符测试 🌍"

		argFile, _ := vm.ToValue(testFile)
		argContent, _ := vm.ToValue(testContent)
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{argFile, argContent},
		}

		result := appendFile(call)

		// appendFile returns null on success
		if !result.IsNull() {
			t.Error("expected null return value for successful append")
		}

		// Verify file was created with correct content
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("failed to read appended file: %v", err)
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

	t.Run("append to existing file", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "existing_append.txt")
		initialContent := "Initial content\n"
		appendContent := "Appended content\n"
		expectedContent := initialContent + appendContent

		// Create initial file
		os.WriteFile(testFile, []byte(initialContent), 0644)

		argFile, _ := vm.ToValue(testFile)
		argContent, _ := vm.ToValue(appendContent)
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{argFile, argContent},
		}

		result := appendFile(call)

		if !result.IsNull() {
			t.Error("expected null return value for successful append")
		}

		// Verify content was appended
		content, _ := os.ReadFile(testFile)
		if string(content) != expectedContent {
			t.Errorf("expected content %q, got %q", expectedContent, string(content))
		}
	})

	t.Run("multiple appends", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "multi_append.txt")
		contents := []string{"Line 1\n", "Line 2\n", "Line 3\n"}
		expectedContent := strings.Join(contents, "")

		argFile, _ := vm.ToValue(testFile)

		// Append multiple times
		for _, content := range contents {
			argContent, _ := vm.ToValue(content)
			call := otto.FunctionCall{
				ArgumentList: []otto.Value{argFile, argContent},
			}

			result := appendFile(call)
			if !result.IsNull() {
				t.Errorf("expected null return value for append of %q", content)
			}
		}

		// Verify all content was appended
		finalContent, _ := os.ReadFile(testFile)
		if string(finalContent) != expectedContent {
			t.Errorf("expected content %q, got %q", expectedContent, string(finalContent))
		}
	})

	t.Run("append to non-existent directory", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "nonexistent", "subdir", "file.txt")
		testContent := "test"

		argFile, _ := vm.ToValue(testFile)
		argContent, _ := vm.ToValue(testContent)
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{argFile, argContent},
		}

		result := appendFile(call)

		// Should return undefined (error)
		if !result.IsUndefined() {
			t.Error("expected undefined when appending to non-existent directory")
		}
	})

	t.Run("append empty content", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "empty_append.txt")
		initialContent := "Initial content"

		// Create initial file
		os.WriteFile(testFile, []byte(initialContent), 0644)

		argFile, _ := vm.ToValue(testFile)
		argContent, _ := vm.ToValue("")
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{argFile, argContent},
		}

		result := appendFile(call)

		if !result.IsNull() {
			t.Error("expected null return value for successful append")
		}

		// Verify content unchanged (empty append)
		content, _ := os.ReadFile(testFile)
		if string(content) != initialContent {
			t.Errorf("expected content %q, got %q", initialContent, string(content))
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

				result := appendFile(call)

				// Should return undefined (error)
				if !result.IsUndefined() {
					t.Error("expected undefined for invalid arguments")
				}
			})
		}
	})

	t.Run("append binary content", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "binary_append.bin")
		initialContent := []byte{0, 1, 2, 3}
		appendContent := string([]byte{255, 254, 253, 252})
		expectedContent := string(initialContent) + appendContent

		// Create initial file with binary content
		os.WriteFile(testFile, initialContent, 0644)

		argFile, _ := vm.ToValue(testFile)
		argContent, _ := vm.ToValue(appendContent)
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{argFile, argContent},
		}

		result := appendFile(call)

		if !result.IsNull() {
			t.Error("expected null return value for successful append")
		}

		// Verify binary content was appended correctly
		content, _ := os.ReadFile(testFile)
		if string(content) != expectedContent {
			t.Error("binary content append mismatch")
		}
	})

	t.Run("append to read-only file", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping read-only test on Windows")
		}

		testFile := filepath.Join(tmpDir, "readonly.txt")
		initialContent := "Initial content\n"

		// Create file and make it read-only
		os.WriteFile(testFile, []byte(initialContent), 0644)
		os.Chmod(testFile, 0444)       // read-only
		defer os.Chmod(testFile, 0644) // restore for cleanup

		argFile, _ := vm.ToValue(testFile)
		argContent, _ := vm.ToValue("This should fail")
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{argFile, argContent},
		}

		result := appendFile(call)

		// Should return undefined (error)
		if !result.IsUndefined() {
			t.Error("expected undefined when appending to read-only file")
		}
	})
}

func TestMkdirAll(t *testing.T) {
	vm := otto.New()

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "js_test_mkdirall_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("create single directory", func(t *testing.T) {
		testDir := filepath.Join(tmpDir, "single")

		arg, _ := vm.ToValue(testDir)
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{arg},
		}

		result := mkdirAll(call)

		// mkdirAll returns null on success
		if !result.IsNull() {
			t.Error("expected null return value for successful directory creation")
		}

		// Verify directory was created
		info, err := os.Stat(testDir)
		if err != nil {
			t.Fatalf("directory was not created: %v", err)
		}

		if !info.IsDir() {
			t.Error("expected directory, got file")
		}

		// Check permissions
		if runtime.GOOS != "windows" {
			if info.Mode().Perm() != 0755 {
				t.Errorf("expected permissions 0755, got %v", info.Mode().Perm())
			}
		}
	})

	t.Run("create nested directories", func(t *testing.T) {
		testDir := filepath.Join(tmpDir, "nested", "sub", "directories")

		arg, _ := vm.ToValue(testDir)
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{arg},
		}

		result := mkdirAll(call)

		if !result.IsNull() {
			t.Error("expected null return value for successful nested directory creation")
		}

		// Verify all directories in the path were created
		currentPath := tmpDir
		for _, part := range []string{"nested", "sub", "directories"} {
			currentPath = filepath.Join(currentPath, part)
			info, err := os.Stat(currentPath)
			if err != nil {
				t.Fatalf("directory %s was not created: %v", currentPath, err)
			}
			if !info.IsDir() {
				t.Errorf("expected %s to be a directory", currentPath)
			}
		}
	})

	t.Run("create existing directory", func(t *testing.T) {
		testDir := filepath.Join(tmpDir, "existing")

		// Create directory first
		os.Mkdir(testDir, 0755)

		arg, _ := vm.ToValue(testDir)
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{arg},
		}

		result := mkdirAll(call)

		// Should succeed (mkdirAll is idempotent)
		if !result.IsNull() {
			t.Error("expected null return value when creating existing directory")
		}

		// Verify directory still exists
		info, err := os.Stat(testDir)
		if err != nil {
			t.Fatalf("existing directory check failed: %v", err)
		}
		if !info.IsDir() {
			t.Error("expected directory to still exist")
		}
	})

	t.Run("create with file in path", func(t *testing.T) {
		// Create a file that will block directory creation
		blockingFile := filepath.Join(tmpDir, "blocking_file.txt")
		os.WriteFile(blockingFile, []byte("blocking"), 0644)

		testDir := filepath.Join(blockingFile, "subdir")

		arg, _ := vm.ToValue(testDir)
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{arg},
		}

		result := mkdirAll(call)

		// Should return undefined (error)
		if !result.IsUndefined() {
			t.Error("expected undefined when file blocks directory creation")
		}
	})

	t.Run("create in read-only directory", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping read-only test on Windows")
		}

		readOnlyDir := filepath.Join(tmpDir, "readonly")
		os.Mkdir(readOnlyDir, 0755)
		os.Chmod(readOnlyDir, 0555)       // read-only
		defer os.Chmod(readOnlyDir, 0755) // restore for cleanup

		testDir := filepath.Join(readOnlyDir, "should_fail")

		arg, _ := vm.ToValue(testDir)
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{arg},
		}

		result := mkdirAll(call)

		// Should return undefined (error)
		if !result.IsUndefined() {
			t.Error("expected undefined when creating directory in read-only parent")
		}
	})

	t.Run("create with special characters", func(t *testing.T) {
		testDir := filepath.Join(tmpDir, "special-chars_123", "with.dots", "and spaces")

		arg, _ := vm.ToValue(testDir)
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{arg},
		}

		result := mkdirAll(call)

		if !result.IsNull() {
			t.Error("expected null return value for directory with special characters")
		}

		// Verify directory was created
		info, err := os.Stat(testDir)
		if err != nil {
			t.Fatalf("directory with special characters was not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("expected directory")
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
					arg1, _ := vm.ToValue("/some/path")
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

				result := mkdirAll(call)

				// Should return undefined (error)
				if !result.IsUndefined() {
					t.Error("expected undefined for invalid arguments")
				}
			})
		}
	})
}

func TestFileSystemIntegration(t *testing.T) {
	vm := otto.New()

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "js_test_integration_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("write, append, then read file", func(t *testing.T) {
		testDir := filepath.Join(tmpDir, "nested", "sub", "directories")
		testFile := filepath.Join(testDir, "roundtrip.txt")
		initialContent := "Round-trip test content\nLine 2\nLine 3\n"
		appendedContent := "Appended content\n"
		expectedContent := initialContent + appendedContent

		// Create subdirectories
		argDir, _ := vm.ToValue(testDir)
		mkdirCall := otto.FunctionCall{
			ArgumentList: []otto.Value{argDir},
		}
		mkdirResult := mkdirAll(mkdirCall)

		if !mkdirResult.IsNull() {
			t.Error("mkdirAll failed")
		}

		// Write file
		argFile, _ := vm.ToValue(testFile)
		argInitial, _ := vm.ToValue(initialContent)
		writeCall := otto.FunctionCall{
			ArgumentList: []otto.Value{argFile, argInitial},
		}

		writeResult := writeFile(writeCall)
		if !writeResult.IsNull() {
			t.Fatal("write failed")
		}

		// Append content
		argAppend, _ := vm.ToValue(appendedContent)
		appendCall := otto.FunctionCall{
			ArgumentList: []otto.Value{argFile, argAppend},
		}

		appendResult := appendFile(appendCall)
		if !appendResult.IsNull() {
			t.Fatal("append failed")
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
		if readContent != expectedContent {
			t.Errorf("round-trip failed: expected %q, got %q", expectedContent, readContent)
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
	tmpFile, _ := os.CreateTemp("", "bench_readfile_*")
	defer os.Remove(tmpFile.Name())

	content := strings.Repeat("Benchmark test content line\n", 100)
	os.WriteFile(tmpFile.Name(), []byte(content), 0644)

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

	tmpDir, _ := os.MkdirTemp("", "bench_writefile_*")
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

func BenchmarkAppendFile(b *testing.B) {
	vm := otto.New()

	tmpDir, _ := os.MkdirTemp("", "bench_appendfile_*")
	defer os.RemoveAll(tmpDir)

	// Create initial file with some content
	testFile := filepath.Join(tmpDir, "bench_append.txt")
	initialContent := "Initial content for benchmark\n"
	os.WriteFile(testFile, []byte(initialContent), 0644)

	content := strings.Repeat("Benchmark append line\n", 10)

	argFile, _ := vm.ToValue(testFile)
	argContent, _ := vm.ToValue(content)
	call := otto.FunctionCall{
		ArgumentList: []otto.Value{argFile, argContent},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = appendFile(call)
	}
}

func BenchmarkMkdirAll(b *testing.B) {
	vm := otto.New()

	tmpDir, _ := os.MkdirTemp("", "bench_mkdirall_*")
	defer os.RemoveAll(tmpDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testDir := filepath.Join(tmpDir, fmt.Sprintf("bench_%d", i), "nested", "sub", "directories")
		arg, _ := vm.ToValue(testDir)
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{arg},
		}
		_ = mkdirAll(call)
	}
}

func BenchmarkReadDir(b *testing.B) {
	vm := otto.New()

	// Create test directory with files
	tmpDir, _ := os.MkdirTemp("", "bench_readdir_*")
	defer os.RemoveAll(tmpDir)

	// Create 100 files
	for i := 0; i < 100; i++ {
		name := filepath.Join(tmpDir, fmt.Sprintf("file_%d.txt", i))
		os.WriteFile(name, []byte("test"), 0644)
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
