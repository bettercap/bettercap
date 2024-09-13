package core

import (
	"os/exec"
	"sort"

	"github.com/bettercap/bettercap/v2/log"
	"github.com/evilsocket/islazy/str"
)

func UniqueInts(a []int, sorted bool) []int {
	tmp := make(map[int]bool, len(a))
	uniq := make([]int, 0, len(a))

	for _, n := range a {
		tmp[n] = true
	}

	for n := range tmp {
		uniq = append(uniq, n)
	}

	if sorted {
		sort.Ints(uniq)
	}

	return uniq
}

func HasBinary(executable string) bool {
	if path, err := exec.LookPath(executable); err != nil || path == "" {
		return false
	}
	return true
}

func Exec(executable string, args []string) (string, error) {
	path, err := exec.LookPath(executable)
	if err != nil {
		log.Warning("executable %s not found in $PATH", executable)
		return "", err
	}

	raw, err := exec.Command(path, args...).CombinedOutput()

	log.Debug("exec=%s args=%v ret_err=%v ret_out=%s", path, args, err, string(raw))
	if err != nil {
		return str.Trim(string(raw)), err
	} else {
		return str.Trim(string(raw)), nil
	}
}
