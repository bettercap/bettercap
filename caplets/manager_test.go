package caplets

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
)

func createTestCaplet(t testing.TB, dir string, name string, content []string) string {
	filename := filepath.Join(dir, name)
	data := strings.Join(content, "\n")
	err := ioutil.WriteFile(filename, []byte(data), 0644)
	if err != nil {
		t.Fatalf("failed to create test caplet: %v", err)
	}
	return filename
}

func TestList(t *testing.T) {
	// Save original values
	origLoadPaths := LoadPaths
	origCache := cache
	cache = make(map[string]*Caplet)

	// Create temp directories
	tempDir, err := ioutil.TempDir("", "caplets-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create subdirectories
	dir1 := filepath.Join(tempDir, "dir1")
	dir2 := filepath.Join(tempDir, "dir2")
	subdir := filepath.Join(dir1, "subdir")

	os.Mkdir(dir1, 0755)
	os.Mkdir(dir2, 0755)
	os.Mkdir(subdir, 0755)

	// Create test caplets
	createTestCaplet(t, dir1, "test1.cap", []string{"# Test caplet 1", "set test 1"})
	createTestCaplet(t, dir1, "test2.cap", []string{"# Test caplet 2", "set test 2"})
	createTestCaplet(t, dir2, "test3.cap", []string{"# Test caplet 3", "set test 3"})
	createTestCaplet(t, subdir, "nested.cap", []string{"# Nested caplet", "set nested test"})

	// Also create a non-caplet file
	ioutil.WriteFile(filepath.Join(dir1, "notacaplet.txt"), []byte("not a caplet"), 0644)

	// Set LoadPaths
	LoadPaths = []string{dir1, dir2}

	// Call List()
	caplets := List()

	// Check results
	if len(caplets) != 4 {
		t.Errorf("expected 4 caplets, got %d", len(caplets))
	}

	// Check names (should be sorted)
	expectedNames := []string{filepath.Join("subdir", "nested"), "test1", "test2", "test3"}
	sort.Strings(expectedNames)

	gotNames := make([]string, len(caplets))
	for i, cap := range caplets {
		gotNames[i] = cap.Name
	}

	for i, expected := range expectedNames {
		if i >= len(gotNames) || gotNames[i] != expected {
			t.Errorf("expected caplet %d to be %s, got %s", i, expected, gotNames[i])
		}
	}

	// Restore original values
	LoadPaths = origLoadPaths
	cache = origCache
}

func TestListEmptyDirectories(t *testing.T) {
	// Save original values
	origLoadPaths := LoadPaths
	origCache := cache
	cache = make(map[string]*Caplet)

	// Create temp directory
	tempDir, err := ioutil.TempDir("", "caplets-empty-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Set LoadPaths to empty directory
	LoadPaths = []string{tempDir}

	// Call List()
	caplets := List()

	// Should return empty list
	if len(caplets) != 0 {
		t.Errorf("expected 0 caplets, got %d", len(caplets))
	}

	// Restore original values
	LoadPaths = origLoadPaths
	cache = origCache
}

func TestLoad(t *testing.T) {
	// Save original values
	origLoadPaths := LoadPaths
	origCache := cache
	cache = make(map[string]*Caplet)

	// Create temp directory
	tempDir, err := ioutil.TempDir("", "caplets-load-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test caplet
	capletContent := []string{
		"# Test caplet",
		"set param value",
		"",
		"# Another comment",
		"run command",
	}
	createTestCaplet(t, tempDir, "test.cap", capletContent)

	// Set LoadPaths
	LoadPaths = []string{tempDir}

	// Test loading without .cap extension
	cap, err := Load("test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cap == nil {
		t.Error("caplet is nil")
	} else {
		if cap.Name != "test" {
			t.Errorf("expected name 'test', got %s", cap.Name)
		}
		if len(cap.Code) != len(capletContent) {
			t.Errorf("expected %d lines, got %d", len(capletContent), len(cap.Code))
		}
	}

	// Test loading from cache
	// Note: The Load function caches with the suffix, so we need to use the same name with suffix
	cap2, err := Load("test.cap")
	if err != nil {
		t.Errorf("unexpected error on cache hit: %v", err)
	}
	if cap2 == nil {
		t.Error("caplet is nil on cache hit")
	}

	// Test loading with .cap extension
	// Note: Load caches by the name parameter, so "test.cap" is a different cache key
	cap3, err := Load("test.cap")
	if err != nil {
		t.Errorf("unexpected error with .cap extension: %v", err)
	}
	if cap3 == nil {
		t.Error("caplet is nil")
	}

	// Restore original values
	LoadPaths = origLoadPaths
	cache = origCache
}

func TestLoadAbsolutePath(t *testing.T) {
	// Save original values
	origCache := cache
	cache = make(map[string]*Caplet)

	// Create temp file
	tempFile, err := ioutil.TempFile("", "test-absolute-*.cap")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	// Write content
	content := "# Absolute path test\nset test absolute"
	tempFile.WriteString(content)
	tempFile.Close()

	// Load with absolute path
	cap, err := Load(tempFile.Name())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cap == nil {
		t.Error("caplet is nil")
	} else {
		if cap.Path != tempFile.Name() {
			t.Errorf("expected path %s, got %s", tempFile.Name(), cap.Path)
		}
	}

	// Restore original values
	cache = origCache
}

func TestLoadNotFound(t *testing.T) {
	// Save original values
	origLoadPaths := LoadPaths
	origCache := cache
	cache = make(map[string]*Caplet)

	// Set empty LoadPaths
	LoadPaths = []string{}

	// Try to load non-existent caplet
	cap, err := Load("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent caplet")
	}
	if cap != nil {
		t.Error("expected nil caplet for non-existent file")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}

	// Restore original values
	LoadPaths = origLoadPaths
	cache = origCache
}

func TestLoadWithFolder(t *testing.T) {
	// Save original values
	origLoadPaths := LoadPaths
	origCache := cache
	cache = make(map[string]*Caplet)

	// Create temp directory structure
	tempDir, err := ioutil.TempDir("", "caplets-folder-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a caplet folder
	capletDir := filepath.Join(tempDir, "mycaplet")
	os.Mkdir(capletDir, 0755)

	// Create main caplet file
	mainContent := []string{"# Main caplet", "set main test"}
	createTestCaplet(t, capletDir, "mycaplet.cap", mainContent)

	// Create additional files
	jsContent := []string{"// JavaScript file", "console.log('test');"}
	createTestCaplet(t, capletDir, "script.js", jsContent)

	capContent := []string{"# Sub caplet", "set sub test"}
	createTestCaplet(t, capletDir, "sub.cap", capContent)

	// Create a file that should be ignored
	ioutil.WriteFile(filepath.Join(capletDir, "readme.txt"), []byte("readme"), 0644)

	// Set LoadPaths
	LoadPaths = []string{tempDir}

	// Load the caplet
	cap, err := Load("mycaplet/mycaplet")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cap == nil {
		t.Fatal("caplet is nil")
	}

	// Check main caplet
	if cap.Name != "mycaplet/mycaplet" {
		t.Errorf("expected name 'mycaplet/mycaplet', got %s", cap.Name)
	}
	if len(cap.Code) != len(mainContent) {
		t.Errorf("expected %d lines in main, got %d", len(mainContent), len(cap.Code))
	}

	// Check additional scripts
	if len(cap.Scripts) != 2 {
		t.Errorf("expected 2 additional scripts, got %d", len(cap.Scripts))
	}

	// Find and check the .js file
	foundJS := false
	foundCap := false
	for _, script := range cap.Scripts {
		if strings.HasSuffix(script.Path, "script.js") {
			foundJS = true
			if len(script.Code) != len(jsContent) {
				t.Errorf("expected %d lines in JS, got %d", len(jsContent), len(script.Code))
			}
		}
		if strings.HasSuffix(script.Path, "sub.cap") {
			foundCap = true
			if len(script.Code) != len(capContent) {
				t.Errorf("expected %d lines in sub.cap, got %d", len(capContent), len(script.Code))
			}
		}
	}

	if !foundJS {
		t.Error("script.js not found in Scripts")
	}
	if !foundCap {
		t.Error("sub.cap not found in Scripts")
	}

	// Restore original values
	LoadPaths = origLoadPaths
	cache = origCache
}

func TestCacheConcurrency(t *testing.T) {
	// Save original values
	origLoadPaths := LoadPaths
	origCache := cache
	cache = make(map[string]*Caplet)

	// Create temp directory
	tempDir, err := ioutil.TempDir("", "caplets-concurrent-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test caplets
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("test%d.cap", i)
		content := []string{fmt.Sprintf("# Test %d", i)}
		createTestCaplet(t, tempDir, name, content)
	}

	// Set LoadPaths
	LoadPaths = []string{tempDir}

	// Run concurrent loads
	var wg sync.WaitGroup
	errors := make(chan error, 50)

	for i := 0; i < 10; i++ {
		for j := 0; j < 5; j++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				name := fmt.Sprintf("test%d", idx)
				_, err := Load(name)
				if err != nil {
					errors <- err
				}
			}(j)
		}
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("concurrent load error: %v", err)
	}

	// Verify cache has all entries
	if len(cache) != 5 {
		t.Errorf("expected 5 cached entries, got %d", len(cache))
	}

	// Restore original values
	LoadPaths = origLoadPaths
	cache = origCache
}

func TestLoadPathPriority(t *testing.T) {
	// Save original values
	origLoadPaths := LoadPaths
	origCache := cache
	cache = make(map[string]*Caplet)

	// Create temp directories
	tempDir1, _ := ioutil.TempDir("", "caplets-priority1-")
	tempDir2, _ := ioutil.TempDir("", "caplets-priority2-")
	defer os.RemoveAll(tempDir1)
	defer os.RemoveAll(tempDir2)

	// Create same-named caplet in both directories
	createTestCaplet(t, tempDir1, "test.cap", []string{"# From dir1"})
	createTestCaplet(t, tempDir2, "test.cap", []string{"# From dir2"})

	// Set LoadPaths with tempDir1 first
	LoadPaths = []string{tempDir1, tempDir2}

	// Load caplet
	cap, err := Load("test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should load from first directory
	if cap != nil && len(cap.Code) > 0 {
		if cap.Code[0] != "# From dir1" {
			t.Error("caplet not loaded from first directory in LoadPaths")
		}
	}

	// Restore original values
	LoadPaths = origLoadPaths
	cache = origCache
}

func BenchmarkLoad(b *testing.B) {
	// Save original values
	origLoadPaths := LoadPaths
	origCache := cache

	// Create temp directory
	tempDir, _ := ioutil.TempDir("", "caplets-bench-")
	defer os.RemoveAll(tempDir)

	// Create test caplet
	content := make([]string, 100)
	for i := range content {
		content[i] = fmt.Sprintf("command %d", i)
	}
	createTestCaplet(b, tempDir, "bench.cap", content)

	// Set LoadPaths
	LoadPaths = []string{tempDir}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear cache to measure loading time
		cache = make(map[string]*Caplet)
		Load("bench")
	}

	// Restore original values
	LoadPaths = origLoadPaths
	cache = origCache
}

func BenchmarkLoadFromCache(b *testing.B) {
	// Save original values
	origLoadPaths := LoadPaths
	origCache := cache
	cache = make(map[string]*Caplet)

	// Create temp directory
	tempDir, _ := ioutil.TempDir("", "caplets-bench-cache-")
	defer os.RemoveAll(tempDir)

	// Create test caplet
	createTestCaplet(b, tempDir, "bench.cap", []string{"# Benchmark"})

	// Set LoadPaths
	LoadPaths = []string{tempDir}

	// Pre-load into cache
	Load("bench")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Load("bench")
	}

	// Restore original values
	LoadPaths = origLoadPaths
	cache = origCache
}

func BenchmarkList(b *testing.B) {
	// Save original values
	origLoadPaths := LoadPaths
	origCache := cache

	// Create temp directory
	tempDir, _ := ioutil.TempDir("", "caplets-bench-list-")
	defer os.RemoveAll(tempDir)

	// Create multiple caplets
	for i := 0; i < 20; i++ {
		name := fmt.Sprintf("test%d.cap", i)
		createTestCaplet(b, tempDir, name, []string{fmt.Sprintf("# Test %d", i)})
	}

	// Set LoadPaths
	LoadPaths = []string{tempDir}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache = make(map[string]*Caplet)
		List()
	}

	// Restore original values
	LoadPaths = origLoadPaths
	cache = origCache
}
