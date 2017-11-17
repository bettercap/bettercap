package core

import (
	"github.com/op/go-logging"
	"os/exec"
	"strings"
)

var log = logging.MustGetLogger("mitm")

func Exec(executable string, args []string) (string, error) {
	path, err := exec.LookPath(executable)
	if err != nil {
		return "", err
	}

	log.Debugf(DIM+"Exec( '%s %s' )"+RESET+"\n", path, strings.Join(args, " "))
	raw, err := exec.Command(path, args...).CombinedOutput()
	if err != nil {
		log.Errorf("  err=%s out='%s'\n", err, raw)
		return "", err
	} else {
		return strings.Trim(string(raw), "\r\n\t "), nil
	}
}
