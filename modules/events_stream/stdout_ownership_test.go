package events_stream

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/bettercap/bettercap/v2/session"
)

func newTestEventsStream(t *testing.T) *EventsStream {
	t.Helper()
	env, err := session.NewEnvironment("")
	if err != nil {
		t.Fatal(err)
	}
	sess := &session.Session{
		Env:              env,
		Events:           session.NewEventPool(false, false),
		EventsIgnoreList: session.NewEventsIgnoreList(),
	}
	// mod.started/mod.stopped would make the worker call Session.Refresh,
	// which needs the interactive prompt the minimal test session lacks
	sess.EventsIgnoreList.Add("mod.")
	return NewEventsStream(sess)
}

// Regression test for https://github.com/bettercap/bettercap/issues/1270
// The API command route (runSessionCommand) temporarily swaps os.Stdout for a
// capture pipe. Stopping events.stream in that window must not close the real
// process stdout.
func TestStopUnderSwappedStdoutDoesNotCloseRealStdout(t *testing.T) {
	mod := newTestEventsStream(t)
	mod.Session.Env.Set("events.stream.output", "")
	mod.Session.Env.Set("events.stream.ignore", "sys.log")
	if err := mod.Start(); err != nil {
		t.Fatal(err)
	}

	real := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	os.Stdout = w // simulate runSessionCommand's capture pipe

	if err := mod.Stop(); err != nil {
		os.Stdout = real
		t.Fatal(err)
	}
	os.Stdout = real

	if _, err := real.Write([]byte("stdout alive\n")); err != nil {
		t.Fatalf("events.stream off closed the real stdout: %v", err)
	}
}

// While the API route has os.Stdout swapped, background event writes must
// keep going to the stdout captured at configure time (baseline behavior),
// never into the API capture pipe.
func TestBackgroundEventsBypassCapturePipe(t *testing.T) {
	realR, realW, err := os.Pipe() // stands in for the process stdout
	if err != nil {
		t.Fatal(err)
	}
	defer realR.Close()
	defer realW.Close()

	oldStdout := os.Stdout
	os.Stdout = realW
	mod := newTestEventsStream(t)
	mod.Session.Env.Set("events.stream.output", "")
	if err := mod.Configure(); err != nil {
		os.Stdout = oldStdout
		t.Fatal(err)
	}

	capR, capW, err := os.Pipe() // the API capture pipe
	if err != nil {
		t.Fatal(err)
	}
	defer capR.Close()
	defer capW.Close()
	os.Stdout = capW // runSessionCommand's swap

	mod.View(session.Event{Tag: "test.event", Data: "hello"}, false)

	os.Stdout = oldStdout

	buf := make([]byte, 4096)
	n, _ := realR.Read(buf)
	if !strings.Contains(string(buf[:n]), "test.event") {
		t.Fatalf("event did not reach the real stdout, got %q", buf[:n])
	}
	capW.Close()
	capBuf := make([]byte, 4096)
	n2, _ := capR.Read(capBuf)
	if n2 != 0 {
		t.Fatalf("event leaked into the API capture pipe: %q", capBuf[:n2])
	}
}

// A module-configured output file must be closed on Stop.
func TestStopClosesOwnedOutputFile(t *testing.T) {
	mod := newTestEventsStream(t)
	logFile := filepath.Join(t.TempDir(), "events.log")
	mod.Session.Env.Set("events.stream.output", logFile)
	mod.Session.Env.Set("events.stream.ignore", "sys.log")
	if err := mod.Start(); err != nil {
		t.Fatal(err)
	}
	fp := mod.outputFile.Load()
	if fp == nil {
		t.Fatal("expected owned output file")
	}
	if err := mod.Stop(); err != nil {
		t.Fatal(err)
	}
	if mod.outputFile.Load() != nil {
		t.Fatal("expected output file reference to be released")
	}
	if _, err := fp.WriteString("x"); err == nil {
		t.Fatal("owned output file was not closed by Stop")
	}
}

// Log rotation must only apply to a module-owned output file and must hand
// over to a fresh file, without ever touching os.Stdout.
func TestRotationOnlyForOwnedFile(t *testing.T) {
	// stdout mode: rotation is a no-op
	mod := newTestEventsStream(t)
	mod.rotation.Enabled = true
	mod.doRotation()
	if mod.outputFile.Load() != nil {
		t.Fatal("rotation must not open a file in stdout mode")
	}

	// file mode: size-based rotation renames, closes the old handle, reopens
	mod = newTestEventsStream(t)
	logFile := filepath.Join(t.TempDir(), "events.log")
	mod.Session.Env.Set("events.stream.output", logFile)
	if err := mod.Configure(); err != nil {
		t.Fatal(err)
	}
	fp := mod.outputFile.Load()
	if fp == nil {
		t.Fatal("expected owned output file")
	}
	if _, err := fp.WriteString("some log data\n"); err != nil {
		t.Fatal(err)
	}
	mod.rotation.Enabled = true
	mod.rotation.How = "size"
	mod.rotation.Period = 0 // any size triggers
	mod.rotation.Format = "20060102150405"
	mod.doRotation()

	matches, err := filepath.Glob(logFile + "-*")
	if err != nil || len(matches) != 1 {
		t.Fatalf("expected one rotated file, got %v (err=%v)", matches, err)
	}
	if _, err := fp.WriteString("x"); err == nil {
		t.Fatal("old output file handle was not closed by rotation")
	}
	fp2 := mod.outputFile.Load()
	if fp2 == nil || fp2 == fp {
		t.Fatal("expected a fresh output file after rotation")
	}
	if _, err := fp2.WriteString("after rotation\n"); err != nil {
		t.Fatalf("writing to rotated file: %v", err)
	}
	if _, err := os.Stat(logFile); err != nil {
		t.Fatalf("expected reopened log file: %v", err)
	}
}

// The ownership pointer is accessed concurrently by Stop (Swap), log
// rotation (Load/Store) and the event worker (Load); this test races those
// accesses against each other and must run clean under -race. It covers the
// ownership pointer only: the pre-existing baseline race on the mod.output
// writer field (rotation reassigns it while the worker reads it) is
// unchanged by this fix and is not exercised here.
func TestOwnershipFieldConcurrentAccess(t *testing.T) {
	mod := newTestEventsStream(t)
	logFile := filepath.Join(t.TempDir(), "events.log")
	mod.Session.Env.Set("events.stream.output", logFile)
	if err := mod.Configure(); err != nil {
		t.Fatal(err)
	}
	mod.rotation.Enabled = true
	mod.rotation.How = "size"
	mod.rotation.Period = 0
	mod.rotation.Format = "20060102150405"

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = mod.outputFile.Load()
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				mod.doRotation()
			}
		}()
	}
	// Stop racing rotation, not serialized after it
	for j := 0; j < 20; j++ {
		if fp := mod.outputFile.Swap(nil); fp != nil {
			fp.Close()
			nf, err := os.CreateTemp(t.TempDir(), "out")
			if err == nil {
				mod.outputFile.Store(nf)
			}
		}
	}
	wg.Wait()
	if fp := mod.outputFile.Swap(nil); fp != nil {
		fp.Close()
	}
}

// View with refresh=true while os.Stdout is swapped must behave exactly like
// baseline: the refresh is skipped (cached writer != current global) and the
// event goes to the configure-time stdout. This is the actual async worker
// path (the worker always calls View with refresh=true); on a minimal test
// session a real Refresh would panic, so a regression here fails loudly.
func TestViewDuringSwapSkipsRefreshLikeBaseline(t *testing.T) {
	realR, realW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer realR.Close()
	defer realW.Close()

	oldStdout := os.Stdout
	os.Stdout = realW
	mod := newTestEventsStream(t)
	mod.Session.Env.Set("events.stream.output", "")
	if err := mod.Configure(); err != nil {
		os.Stdout = oldStdout
		t.Fatal(err)
	}

	capR, capW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer capR.Close()
	defer capW.Close()
	os.Stdout = capW

	// must not panic via Session.Refresh on the minimal test session
	mod.View(session.Event{Tag: "test.event", Data: "hello"}, true)

	os.Stdout = oldStdout
	buf := make([]byte, 4096)
	n, _ := realR.Read(buf)
	if !strings.Contains(string(buf[:n]), "test.event") {
		t.Fatalf("event did not reach the real stdout, got %q", buf[:n])
	}
}
