package core

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

const (
	defaultTrimSet = "\r\n\t "
)

func Trim(s string) string {
	return strings.Trim(s, defaultTrimSet)
}

func TrimRight(s string) string {
	return strings.TrimRight(s, defaultTrimSet)
}

func Exec(executable string, args []string) (string, error) {
	path, err := exec.LookPath(executable)
	if err != nil {
		return "", err
	}

	raw, err := exec.Command(path, args...).CombinedOutput()
	if err != nil {
		fmt.Printf("ERROR: path=%s args=%s err=%s out='%s'\n", path, args, err, raw)
		return "", err
	} else {
		return Trim(string(raw)), nil
	}
}

func Exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func ExpandPath(path string) (string, error) {
	// Check if path is empty
	if path != "" {
		if strings.HasPrefix(path, "~") {
			usr, err := user.Current()
			if err != nil {
				return "", err
			}
			// Replace only the first occurrence of ~
			path = strings.Replace(path, "~", usr.HomeDir, 1)
		}
		return filepath.Abs(path)
	}
	return "", nil
}
