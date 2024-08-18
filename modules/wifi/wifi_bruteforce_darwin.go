package wifi

import (
	"errors"
	"os/exec"
	"time"

	"github.com/bettercap/bettercap/v2/core"
	"github.com/evilsocket/islazy/async"
)

func wifiBruteforce(mod *WiFiModule, job bruteforceJob) (bool, error) {
	networksetup, err := exec.LookPath("networksetup")
	if err != nil {
		return false, errors.New("could not find networksetup in $PATH")
	}

	args := []string{
		"-setairportnetwork",
		job.iface,
		job.essid,
		job.password,
	}

	type result struct {
		auth bool
		err  error
	}

	if res, err := async.WithTimeout(job.timeout, func() interface{} {
		start := time.Now()
		if output, err := core.Exec(networksetup, args); err != nil {
			return result{auth: false, err: err}
		} else {
			mod.Debug("%s %v : %v\n%v", networksetup, args, time.Since(start), output)
			return result{auth: output == "", err: nil}
		}
	}); err == nil && res != nil {
		return res.(result).auth, res.(result).err
	}

	return false, nil
}
