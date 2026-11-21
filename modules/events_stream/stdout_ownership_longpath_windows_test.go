package events_stream

import (
	"testing"

	"golang.org/x/sys/windows"
)

// tempDir returns t.TempDir() resolved to its long (non 8.3) form.
//
// On the GitHub Windows runners t.TempDir() sits under the profile's 8.3 short
// name, e.g. C:\Users\RUNNER~1\AppData\Local\Temp\... . Configure() feeds
// events.stream.output through fs.Expand(), which replaces every '~' with $HOME,
// so the literal '~' in "RUNNER~1" mangles the path and the open fails. Resolving
// to the long path first yields a tilde-free directory, matching real usage.
func tempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	long, err := longPathName(dir)
	if err != nil {
		t.Fatalf("resolving long path of %q: %v", dir, err)
	}
	return long
}

func longPathName(p string) (string, error) {
	u16, err := windows.UTF16PtrFromString(p)
	if err != nil {
		return "", err
	}
	buf := make([]uint16, len(p)+260)
	n, err := windows.GetLongPathName(u16, &buf[0], uint32(len(buf)))
	if err != nil {
		return "", err
	}
	// n > len(buf) means the buffer was too small and n is the required size.
	if int(n) > len(buf) {
		buf = make([]uint16, n)
		n, err = windows.GetLongPathName(u16, &buf[0], uint32(len(buf)))
		if err != nil {
			return "", err
		}
	}
	return windows.UTF16ToString(buf[:n]), nil
}
