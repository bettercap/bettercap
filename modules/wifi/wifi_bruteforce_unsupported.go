//go:build windows || freebsd || netbsd || openbsd
// +build windows freebsd netbsd openbsd

package wifi

import "errors"

func wifiBruteforce(_ *WiFiModule, _ bruteforceJob) (bool, error) {
	return false, errors.New("not supported on this OS")
}
