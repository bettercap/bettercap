package core

import (
	"fmt"
	"os/exec"
	"sort"

	"github.com/evilsocket/islazy/str"
)

func UniqueInts(a []int, sorted bool) []int {
	tmp := make(map[int]bool)
	uniq := make([]int, 0)

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

func ExecSilent(executable string, args []string) (string, error) {
	path, err := exec.LookPath(executable)
	if err != nil {
		return "", err
	}

	raw, err := exec.Command(path, args...).CombinedOutput()
	if err != nil {
		return "", err
	} else {
		return str.Trim(string(raw)), nil
	}
}

func Exec(executable string, args []string) (string, error) {
	out, err := ExecSilent(executable, args)
	if err != nil {
		fmt.Printf("ERROR for '%s %s': %s\n", executable, args, err)
	}
	return out, err
}
