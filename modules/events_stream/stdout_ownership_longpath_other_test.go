//go:build !windows

package events_stream

import "testing"

// tempDir is a plain t.TempDir() everywhere but Windows, where the short-name
// handling in the _windows variant is needed. See that file for details.
func tempDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}
