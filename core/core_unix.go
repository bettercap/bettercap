// +build !windows,!android

package core

func Shell(cmd string) (string, error) {
	return Exec("/bin/sh", []string{"-c", cmd})
}
