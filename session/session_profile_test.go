package session

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bettercap/bettercap/v2/core"
)

func TestCPUProfileIsValidAfterClose(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cpu.prof")
	s, err := WithOptions(core.Options{CpuProfile: path})
	if err != nil {
		t.Fatalf("WithOptions: %v", err)
	}
	defer s.Close()

	deadline := time.Now().Add(300 * time.Millisecond)
	x := 0
	for time.Now().Before(deadline) {
		for i := 0; i < 10000; i++ {
			x += i
		}
	}
	_ = x

	s.Close()

	st, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat profile: %v", err)
	}
	if st.Size() == 0 {
		t.Fatalf("CPU profile %s is empty; profile file was closed before StopCPUProfile", path)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open profile: %v", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("CPU profile is not valid gzip: %v", err)
	}
	defer gz.Close()

	if _, err := io.Copy(io.Discard, gz); err != nil {
		t.Fatalf("CPU profile is truncated: %v", err)
	}
}
