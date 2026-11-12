package api_rest

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/bettercap/bettercap/v2/modules/events_stream"
	"github.com/bettercap/bettercap/v2/session"
)

// countFDs reports the process fd count on Linux. The fd-balance check is
// Linux-only; when /proc is unavailable (macOS, Windows) it returns ok=false
// and the functional assertions of the test still run.
func countFDs() (n int, ok bool) {
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		return 0, false
	}
	return len(entries), true
}

func checkFDBalance(t *testing.T, before int, ok bool) {
	t.Helper()
	if !ok {
		return
	}
	if after, _ := countFDs(); after != before {
		t.Fatalf("capture pipe leaked: %d fds before, %d after", before, after)
	}
}

// A failing API command must still restore os.Stdout and release the capture
// pipe; previously stdout stayed swapped and both pipe ends leaked.
func TestFailedCommandRestoresStdoutAndClosesPipe(t *testing.T) {
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
	mod := &RestAPI{
		SessionModule: session.NewSessionModule("api.rest", sess),
	}

	fdsBefore, fdsOK := countFDs()
	real := os.Stdout
	req := httptest.NewRequest("POST", "/api/session", strings.NewReader(`{"cmd":"no.such.command"}`))
	rec := httptest.NewRecorder()

	mod.runSessionCommand(rec, req)

	if rec.Code != 400 {
		t.Fatalf("expected status 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "unknown or invalid syntax") {
		t.Fatalf("expected the command error in the response body, got %q", rec.Body.String())
	}
	if os.Stdout != real {
		t.Fatal("os.Stdout was left swapped to the capture pipe")
	}
	if _, err := real.Write([]byte("stdout alive\n")); err != nil {
		t.Fatalf("the failing command closed the real stdout: %v", err)
	}
	checkFDBalance(t, fdsBefore, fdsOK)
}

// A successful API command keeps the original behavior: output captured,
// status 200 with a success body.
func TestSuccessfulCommandCapturesOutput(t *testing.T) {
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
	eventsMod := events_stream.NewEventsStream(sess)
	sess.Env.Set("events.stream.ignore", "sys.log")
	if err := sess.Register(eventsMod); err != nil {
		t.Fatal(err)
	}
	mod := &RestAPI{
		SessionModule: session.NewSessionModule("api.rest", sess),
	}

	defer eventsMod.Stop() // do not leave the module worker running

	fdsBefore, fdsOK := countFDs()
	req := httptest.NewRequest("POST", "/api/session", strings.NewReader(`{"cmd":"events.stream on; events.filters"}`))
	rec := httptest.NewRecorder()

	mod.runSessionCommand(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var res APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decoding response: %v (body %q)", err, rec.Body.String())
	}
	if !res.Success {
		t.Fatalf("expected success body, got %s", rec.Body.String())
	}
	if strings.TrimSpace(res.Message) == "" {
		t.Fatal("expected non-empty captured command output in the response")
	}
	if !strings.Contains(res.Message, "mod.") {
		t.Fatalf("expected the events.filters output in the captured message, got %q", res.Message)
	}
	checkFDBalance(t, fdsBefore, fdsOK)
}

// Exact-layer regression for https://github.com/bettercap/bettercap/issues/1270
// Drives the real API command route with the real events.stream module:
// POST {"cmd":"events.stream off"} must stop the stream and leave the
// process stdout fully functional.
func TestEventsStreamOffViaAPIRouteKeepsStdout(t *testing.T) {
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
	eventsMod := events_stream.NewEventsStream(sess)
	sess.Env.Set("events.stream.ignore", "sys.log")
	if err := sess.Register(eventsMod); err != nil {
		t.Fatal(err)
	}
	mod := &RestAPI{
		SessionModule: session.NewSessionModule("api.rest", sess),
	}

	eventsMod.Session.Env.Set("events.stream.ignore", "sys.log")
	if err := eventsMod.Start(); err != nil {
		t.Fatalf("starting events.stream: %v", err)
	}
	defer eventsMod.Stop() // backstop cleanup; the command under test should have stopped it already

	fdsBefore, fdsOK := countFDs()
	real := os.Stdout
	req := httptest.NewRequest("POST", "/api/session", strings.NewReader(`{"cmd":"events.stream off"}`))
	rec := httptest.NewRecorder()

	mod.runSessionCommand(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"success":true`) {
		t.Fatalf("expected success body, got %s", rec.Body.String())
	}
	if eventsMod.Running() {
		t.Fatal("expected events.stream to be stopped")
	}
	if os.Stdout != real {
		t.Fatal("os.Stdout was not restored")
	}
	if _, err := real.Write([]byte("stdout alive\n")); err != nil {
		t.Fatalf("the API route closed the real stdout: %v", err)
	}
	checkFDBalance(t, fdsBefore, fdsOK)
}
