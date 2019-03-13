// +build android

package core

func Shell(cmd string) (string, error) {
	return Exec("/system/bin/sh", []string{"-c", cmd})
}
