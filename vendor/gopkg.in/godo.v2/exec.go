package godo

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/howeyc/gopass"
	"github.com/mgutz/str"
	"github.com/nozzle/throttler"
	"gopkg.in/godo.v2/util"
)

// Bash executes a bash script (string).
func Bash(script string, options ...map[string]interface{}) (string, error) {
	return bash(script, options)
}

// BashOutput executes a bash script and returns the output
func BashOutput(script string, options ...map[string]interface{}) (string, error) {
	if len(options) == 0 {
		options = append(options, M{"$out": CaptureBoth})
	} else {
		options[0]["$out"] = CaptureBoth
	}
	return bash(script, options)
}

// Run runs a command.
func Run(commandstr string, options ...map[string]interface{}) (string, error) {
	return run(commandstr, options)
}

// RunOutput runs a command and returns output.
func RunOutput(commandstr string, options ...map[string]interface{}) (string, error) {
	if len(options) == 0 {
		options = append(options, M{"$out": CaptureBoth})
	} else {
		options[0]["$out"] = CaptureBoth
	}
	return run(commandstr, options)
}

// Start starts an async command. If executable has suffix ".go" then it will
// be "go install"ed then executed. Use this for watching a server task.
//
// If Start is called with the same command it kills the previous process.
//
// The working directory is optional.
func Start(commandstr string, options ...map[string]interface{}) error {
	return startEx(nil, commandstr, options)
}

func rebuildPackage(filename string) error {
	_, err := Run("go build", M{"$in": filepath.Dir(filename)})
	return err
}

func startEx(context *Context, commandstr string, options []map[string]interface{}) error {
	m, dir, _, err := parseOptions(options)
	if err != nil {
		return err
	}
	if strings.Contains(commandstr, "{{") {
		commandstr, err = util.StrTemplate(commandstr, m)
		if err != nil {
			return err
		}
	}
	executable, argv, env := splitCommand(commandstr)
	if context != nil && context.FileEvent != nil {
		event := context.FileEvent
		absPath, err := filepath.Abs(filepath.Join(dir, executable))
		if err != nil {
			return err
		}
		if filepath.Ext(event.Path) == ".go" && event.Path != absPath {
			var p string
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			p, err = filepath.Rel(wd, event.Path)
			if err != nil {
				p = event.Path
			}
			util.Info(context.Task.Name, "rebuilding %s...\n", filepath.Dir(p))
			rebuildPackage(event.Path)
		}
	}
	isGoFile := strings.HasSuffix(executable, ".go")
	if isGoFile {
		cmdstr := "go install"
		if context == nil || context.FileEvent == nil {
			util.Info(context.Task.Name, "rebuilding with -a to ensure clean build (might take awhile)\n")
			cmdstr += " -a"
		}
		_, err = Run(cmdstr, m)
		if err != nil {
			return err
		}
		executable = filepath.Base(dir)
	}
	cmd := &command{
		executable: executable,
		wd:         dir,
		env:        env,
		argv:       argv,
		commandstr: commandstr,
	}
	return cmd.runAsync()
}

func getWorkingDir(m map[string]interface{}) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", nil
	}

	var wd string
	if m != nil {
		if d, ok := m["$in"].(string); ok {
			wd = d
		}
	}
	if wd != "" {
		var path string
		if filepath.IsAbs(wd) {
			path = wd
		} else {
			path = filepath.Join(pwd, wd)
		}
		_, err := os.Stat(path)
		if err == nil {
			return path, nil
		}
		return "", fmt.Errorf("working dir does not exist: %s", path)
	}
	return pwd, nil
}

func parseOptions(options []map[string]interface{}) (m map[string]interface{}, dir string, capture int, err error) {
	if options == nil {
		m = map[string]interface{}{}
	} else {
		m = options[0]
	}

	dir, err = getWorkingDir(m)
	if err != nil {
		return nil, "", 0, err
	}

	if n, ok := m["$out"].(int); ok {
		capture = n
	}

	return m, dir, capture, nil
}

// Bash executes a bash string. Use backticks for multiline. To execute as shell script,
// use Run("bash script.sh")
func bash(script string, options []map[string]interface{}) (output string, err error) {
	m, dir, capture, err := parseOptions(options)
	if err != nil {
		return "", err
	}

	if strings.Contains(script, "{{") {
		script, err = util.StrTemplate(script, m)
		if err != nil {
			return "", err
		}
	}

	gcmd := &command{
		executable: "bash",
		argv:       []string{"-c", script},
		wd:         dir,
		capture:    capture,
		commandstr: script,
	}

	return gcmd.run()
}

func run(commandstr string, options []map[string]interface{}) (output string, err error) {
	m, dir, capture, err := parseOptions(options)
	if err != nil {
		return "", err
	}

	if strings.Contains(commandstr, "{{") {
		commandstr, err = util.StrTemplate(commandstr, m)
		if err != nil {
			return "", err
		}
	}

	lines := strings.Split(commandstr, "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("Empty command string")
	}
	for i, cmdline := range lines {
		cmdstr := strings.Trim(cmdline, " \t")
		if cmdstr == "" {
			continue
		}
		executable, argv, env := splitCommand(cmdstr)

		cmd := &command{
			executable: executable,
			wd:         dir,
			env:        env,
			argv:       argv,
			capture:    capture,
			commandstr: commandstr,
		}

		s, err := cmd.run()
		if err != nil {
			err = fmt.Errorf(err.Error()+"\nline=%d", i)
			return s, err
		}
		output += s
	}
	return output, nil
}

// func getWd(wd []In) (string, error) {
// 	if len(wd) == 1 {
// 		return wd[0][0], nil
// 	}
// 	return os.Getwd()
// }

func splitCommand(command string) (executable string, argv, env []string) {
	argv = str.ToArgv(command)
	for i, item := range argv {
		if strings.Contains(item, "=") {
			if env == nil {
				env = []string{item}
				continue
			}
			env = append(env, item)
		} else {
			executable = item
			argv = argv[i+1:]
			return
		}
	}

	executable = argv[0]
	argv = argv[1:]
	return
}

func toInt(s string) int {
	result, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return result
}

// Inside temporarily changes the working directory and restores it when lambda
// finishes.
func Inside(dir string, lambda func()) error {
	olddir, err := os.Getwd()
	if err != nil {
		return err
	}

	err = os.Chdir(dir)
	if err != nil {
		return err
	}

	defer func() {
		os.Chdir(olddir)
	}()
	lambda()
	return nil
}

// Prompt prompts user for input with default value.
func Prompt(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	text, _ := reader.ReadString('\n')
	return text
}

// PromptPassword prompts user for password input.
func PromptPassword(prompt string) string {
	fmt.Printf(prompt)
	b, err := gopass.GetPasswd()
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return string(b)
}

// GoThrottle starts to run the given list of fns concurrently,
// at most n fns at a time.
func GoThrottle(throttle int, fns ...func() error) error {
	var err error

	// Create a new Throttler that will get 2 urls at a time
	t := throttler.New(throttle, len(fns))
	for _, fn := range fns {
		// Launch a goroutine to fetch the URL.
		go func(f func() error) {
			err2 := f()
			if err2 != nil {
				err = err2
			}

			// Let Throttler know when the goroutine completes
			// so it can dispatch another worker
			t.Done(err)
		}(fn)
		// Pauses until a worker is available or all jobs have been completed
		// Returning the total number of goroutines that have errored
		// lets you choose to break out of the loop without starting any more
		errorCount := t.Throttle()
		if errorCount > 0 {
			break
		}
	}
	return err
}
