package core

func Shell(cmd string) (string, error) {
	return Exec("cmd.exe", []string{"/c", cmd})
}

func ExecInEnglish(executable string, args []string) (string, error) {
	return Exec("cmd", append([]string{"/C", "chcp 437", "&&", executable}, args...))
}