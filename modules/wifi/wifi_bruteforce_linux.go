package wifi

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/islazy/tui"
)

func wifiBruteforce(mod *WiFiModule, job bruteforceJob) (bool, error) {
	wpa_supplicant, err := exec.LookPath("wpa_supplicant")
	if err != nil {
		return false, errors.New("could not find wpa_supplicant in $PATH")
	}

	config := fmt.Sprintf(`p2p_disabled=1 
	network={
		ssid=%s
		psk=%s
	}`, strconv.Quote(job.essid), strconv.Quote(job.password))

	file, err := os.CreateTemp("", "bettercap-wpa-config")
	if err != nil {
		return false, fmt.Errorf("could not create temporary configuration file: %v", err)
	}
	defer os.Remove(file.Name())

	if _, err := file.WriteString(config); err != nil {
		return false, fmt.Errorf("could not write temporary configuration file: %v", err)
	}

	mod.Debug("using %s ...", file.Name())

	args := []string{
		"-i",
		job.iface,
		"-c",
		file.Name(),
	}
	cmd := exec.Command(wpa_supplicant, args...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return false, err
	}

	scanner := bufio.NewScanner(cmdReader)
	done := make(chan bool)
	go func() {
		auth := false
		for scanner.Scan() {
			line := strings.ToLower(str.Trim(scanner.Text()))
			if strings.Contains(line, "handshake failed") {
				mod.Debug("%s", tui.Red(line))
				break
			} else if strings.Contains(line, "key negotiation completed") {
				mod.Debug("%s", tui.Bold(tui.Green(line)))
				auth = true
				break
			} else {
				mod.Debug("%s", tui.Dim(line))
			}
		}
		if auth {
			mod.Debug("success: %v", job)
		}
		done <- auth
	}()

	if err := cmd.Start(); err != nil {
		return false, err
	}

	timeout := time.After(job.timeout)
	doneInTime := make(chan bool)
	go func() {
		doneInTime <- <-done
	}()

	select {
	case <-timeout:
		mod.Debug("%s timeout", job.password)
		// make sure the process is killed
		cmd.Process.Kill()
		return false, nil
	case res := <-doneInTime:
		mod.Debug("%s=%v", job.password, res)
		// make sure the process is killed
		cmd.Process.Kill()
		return res, nil
	}
}
