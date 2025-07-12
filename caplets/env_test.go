package caplets

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetDefaultInstallBase(t *testing.T) {
	base := getDefaultInstallBase()

	if runtime.GOOS == "windows" {
		expected := filepath.Join(os.Getenv("ALLUSERSPROFILE"), "bettercap")
		if base != expected {
			t.Errorf("on windows, expected %s, got %s", expected, base)
		}
	} else {
		expected := "/usr/local/share/bettercap/"
		if base != expected {
			t.Errorf("on non-windows, expected %s, got %s", expected, base)
		}
	}
}

func TestGetUserHomeDir(t *testing.T) {
	home := getUserHomeDir()

	// Should return a non-empty string
	if home == "" {
		t.Error("getUserHomeDir returned empty string")
	}

	// Should be an absolute path
	if !filepath.IsAbs(home) {
		t.Errorf("expected absolute path, got %s", home)
	}
}

func TestSetup(t *testing.T) {
	// Save original values
	origInstallBase := InstallBase
	origInstallPathArchive := InstallPathArchive
	origInstallPath := InstallPath
	origArchivePath := ArchivePath
	origLoadPaths := LoadPaths

	// Test with custom base
	testBase := "/custom/base"
	err := Setup(testBase)

	if err != nil {
		t.Errorf("Setup returned error: %v", err)
	}

	// Check that paths are set correctly
	if InstallBase != testBase {
		t.Errorf("expected InstallBase %s, got %s", testBase, InstallBase)
	}

	expectedArchivePath := filepath.Join(testBase, "caplets-master")
	if InstallPathArchive != expectedArchivePath {
		t.Errorf("expected InstallPathArchive %s, got %s", expectedArchivePath, InstallPathArchive)
	}

	expectedInstallPath := filepath.Join(testBase, "caplets")
	if InstallPath != expectedInstallPath {
		t.Errorf("expected InstallPath %s, got %s", expectedInstallPath, InstallPath)
	}

	expectedTempPath := filepath.Join(os.TempDir(), "caplets.zip")
	if ArchivePath != expectedTempPath {
		t.Errorf("expected ArchivePath %s, got %s", expectedTempPath, ArchivePath)
	}

	// Check LoadPaths contains expected paths
	expectedInLoadPaths := []string{
		"./",
		"./caplets/",
		InstallPath,
		filepath.Join(getUserHomeDir(), "caplets"),
	}

	for _, expected := range expectedInLoadPaths {
		absExpected, _ := filepath.Abs(expected)
		found := false
		for _, path := range LoadPaths {
			if path == absExpected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected path %s not found in LoadPaths", absExpected)
		}
	}

	// All paths should be absolute
	for _, path := range LoadPaths {
		if !filepath.IsAbs(path) {
			t.Errorf("LoadPath %s is not absolute", path)
		}
	}

	// Restore original values
	InstallBase = origInstallBase
	InstallPathArchive = origInstallPathArchive
	InstallPath = origInstallPath
	ArchivePath = origArchivePath
	LoadPaths = origLoadPaths
}

func TestSetupWithEnvironmentVariable(t *testing.T) {
	// Save original values
	origEnv := os.Getenv(EnvVarName)
	origLoadPaths := LoadPaths

	// Set environment variable with multiple paths
	testPaths := []string{"/path1", "/path2", "/path3"}
	os.Setenv(EnvVarName, strings.Join(testPaths, string(os.PathListSeparator)))

	// Run setup
	err := Setup("/test/base")
	if err != nil {
		t.Errorf("Setup returned error: %v", err)
	}

	// Check that custom paths from env var are in LoadPaths
	for _, testPath := range testPaths {
		absTestPath, _ := filepath.Abs(testPath)
		found := false
		for _, path := range LoadPaths {
			if path == absTestPath {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected env path %s not found in LoadPaths", absTestPath)
		}
	}

	// Restore original values
	if origEnv == "" {
		os.Unsetenv(EnvVarName)
	} else {
		os.Setenv(EnvVarName, origEnv)
	}
	LoadPaths = origLoadPaths
}

func TestSetupWithEmptyEnvironmentVariable(t *testing.T) {
	// Save original values
	origEnv := os.Getenv(EnvVarName)
	origLoadPaths := LoadPaths

	// Set empty environment variable
	os.Setenv(EnvVarName, "")

	// Count LoadPaths before setup
	err := Setup("/test/base")
	if err != nil {
		t.Errorf("Setup returned error: %v", err)
	}

	// Should have only the default paths (4)
	if len(LoadPaths) != 4 {
		t.Errorf("expected 4 default LoadPaths, got %d", len(LoadPaths))
	}

	// Restore original values
	if origEnv == "" {
		os.Unsetenv(EnvVarName)
	} else {
		os.Setenv(EnvVarName, origEnv)
	}
	LoadPaths = origLoadPaths
}

func TestSetupWithWhitespaceInEnvironmentVariable(t *testing.T) {
	// Save original values
	origEnv := os.Getenv(EnvVarName)
	origLoadPaths := LoadPaths

	// Set environment variable with whitespace
	testPaths := []string{"  /path1  ", "  ", "/path2  "}
	os.Setenv(EnvVarName, strings.Join(testPaths, string(os.PathListSeparator)))

	// Run setup
	err := Setup("/test/base")
	if err != nil {
		t.Errorf("Setup returned error: %v", err)
	}

	// Should have added only non-empty paths after trimming
	expectedPaths := []string{"/path1", "/path2"}
	foundCount := 0
	for _, expectedPath := range expectedPaths {
		absExpected, _ := filepath.Abs(expectedPath)
		for _, path := range LoadPaths {
			if path == absExpected {
				foundCount++
				break
			}
		}
	}

	if foundCount != len(expectedPaths) {
		t.Errorf("expected to find %d paths from env, found %d", len(expectedPaths), foundCount)
	}

	// Restore original values
	if origEnv == "" {
		os.Unsetenv(EnvVarName)
	} else {
		os.Setenv(EnvVarName, origEnv)
	}
	LoadPaths = origLoadPaths
}

func TestConstants(t *testing.T) {
	// Test that constants have expected values
	if EnvVarName != "CAPSPATH" {
		t.Errorf("expected EnvVarName to be 'CAPSPATH', got %s", EnvVarName)
	}

	if Suffix != ".cap" {
		t.Errorf("expected Suffix to be '.cap', got %s", Suffix)
	}

	if InstallArchive != "https://github.com/bettercap/caplets/archive/master.zip" {
		t.Errorf("unexpected InstallArchive value: %s", InstallArchive)
	}
}

func TestInit(t *testing.T) {
	// The init function should have been called already
	// Check that paths are initialized
	if InstallBase == "" {
		t.Error("InstallBase not initialized")
	}

	if InstallPath == "" {
		t.Error("InstallPath not initialized")
	}

	if InstallPathArchive == "" {
		t.Error("InstallPathArchive not initialized")
	}

	if ArchivePath == "" {
		t.Error("ArchivePath not initialized")
	}

	if LoadPaths == nil || len(LoadPaths) == 0 {
		t.Error("LoadPaths not initialized")
	}
}

func TestSetupMultipleTimes(t *testing.T) {
	// Save original values
	origLoadPaths := LoadPaths

	// Setup multiple times with different bases
	bases := []string{"/base1", "/base2", "/base3"}

	for _, base := range bases {
		err := Setup(base)
		if err != nil {
			t.Errorf("Setup(%s) returned error: %v", base, err)
		}

		// Check that InstallBase is updated
		if InstallBase != base {
			t.Errorf("expected InstallBase %s, got %s", base, InstallBase)
		}

		// LoadPaths should be recreated each time
		if len(LoadPaths) < 4 {
			t.Errorf("LoadPaths should have at least 4 entries, got %d", len(LoadPaths))
		}
	}

	// Restore original values
	LoadPaths = origLoadPaths
}

func BenchmarkSetup(b *testing.B) {
	// Save original values
	origEnv := os.Getenv(EnvVarName)

	// Set a complex environment
	paths := []string{"/p1", "/p2", "/p3", "/p4", "/p5"}
	os.Setenv(EnvVarName, strings.Join(paths, string(os.PathListSeparator)))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Setup("/benchmark/base")
	}

	// Restore
	if origEnv == "" {
		os.Unsetenv(EnvVarName)
	} else {
		os.Setenv(EnvVarName, origEnv)
	}
}
