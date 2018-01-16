package core

func Shell(cmd string) (string, error) {
	return Exec("cmd.exe", []string{"/c", cmd})
}
