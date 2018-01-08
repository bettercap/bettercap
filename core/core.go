package core

import (
	"fmt"
	"os/exec"
	"strings"
)

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
		return strings.Trim(string(raw), "\r\n\t "), nil
	}
}
