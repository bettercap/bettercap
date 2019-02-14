package godo

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mgutz/minimist"
	"gopkg.in/godo.v2/util"
	"gopkg.in/godo.v2/watcher"
)

// Message are sent on the Events channel
type Message struct {
	Event string
	Data  string
}

const defaultWatchDelay = 1200 * time.Millisecond

var watching bool
var help bool
var verbose bool
var version bool
var deprecatedWarnings bool

// DebounceMs is the default time (1500 ms) to debounce task events in watch mode.
var Debounce time.Duration
var runnerWaitGroup = &WaitGroupN{}
var waitExit bool
var argm minimist.ArgMap
var wd string
var watchDelay = defaultWatchDelay

// SetWatchDelay sets the time duration between watches.
func SetWatchDelay(delay time.Duration) {
	if delay == 0 {
		delay = defaultWatchDelay
	}
	watchDelay = delay
	watcher.SetWatchDelay(watchDelay)
}

// GetWatchDelay gets the watch delay
func GetWatchDelay() time.Duration {
	return watchDelay
}

func init() {
	// WatchDelay is the time to poll the file system
	SetWatchDelay(watchDelay)
	Debounce = 2000 * time.Millisecond
	var err error
	wd, err = os.Getwd()
	if err != nil {
		panic(err)
	}

}

// Usage prints a usage screen with task descriptions.
func Usage(tasks string) {
	// go's flag package prints ugly screen
	format := `godo %s - do task(s)

Usage: godo [flags] [task...]
  -D             Print deprecated warnings
      --dump     Dump debug info about the project
  -h, --help     This screen
  -i, --install  Install Godofile dependencies
      --rebuild  Rebuild Godofile
  -v  --verbose  Log verbosely
  -V, --version  Print version
  -w, --watch    Watch task and dependencies`

	if tasks == "" {
		fmt.Printf(format, Version)
	} else {
		format += "\n\n%s"
		fmt.Printf(format, Version, tasks)
	}
}

// Godo runs a project of tasks.
func Godo(tasksFunc func(*Project)) {
	godo(tasksFunc, nil)
}

func godo(tasksFn func(*Project), argv []string) {
	godoExit(tasksFn, argv, os.Exit)
}

// used for testing to switch out exitFn
func godoExit(tasksFunc func(*Project), argv []string, exitFn func(int)) {
	if argv == nil {
		argm = minimist.Parse()
	} else {
		argm = minimist.ParseArgv(argv)
	}

	dump := argm.AsBool("dump")
	help = argm.AsBool("help", "h", "?")
	verbose = argm.AsBool("verbose", "v")
	version = argm.AsBool("version", "V")
	watching = argm.AsBool("watch", "w")
	deprecatedWarnings = argm.AsBool("D")
	contextArgm := minimist.ParseArgv(argm.Unparsed())

	project := NewProject(tasksFunc, exitFn, contextArgm)

	if help {
		Usage(project.usage())
		exitFn(0)
	}

	if version {
		fmt.Printf("godo %s\n", Version)
		exitFn(0)
	}

	if dump {
		project.dump(os.Stdout, "", "  ")
		exitFn(0)
	}

	// env vars are any nonflag key=value pair
	addToOSEnviron(argm.NonFlags())

	// Run each task including their dependencies.
	args := []string{}
	for _, s := range argm.NonFlags() {
		// skip env vars
		if !strings.Contains(s, "=") {
			args = append(args, s)
		}
	}

	if len(args) == 0 {
		if project.Tasks["default"] != nil {
			args = append(args, "default")
		} else {
			Usage(project.usage())
			exitFn(0)
		}
	}

	for _, name := range args {
		err := project.Run(name)
		if err != nil {
			util.Error("ERR", "%s\n", err.Error())
			exitFn(1)
		}
	}

	if watching {
		if project.Watch(args, true) {
			runnerWaitGroup.Add(1)
			waitExit = true
		} else {
			fmt.Println("Nothing to watch. Use Task#Src() to specify watch patterns")
			exitFn(0)
		}
	}

	if waitExit {
		// Ctrl+C handler
		csig := make(chan os.Signal, 1)
		signal.Notify(csig, syscall.SIGQUIT)
		go func() {
			for sig := range csig {
				fmt.Println("SIG caught")
				if sig == syscall.SIGQUIT {
					fmt.Println("SIG caught B")
					project.Exit(0)
					break
				}
			}
		}()

		runnerWaitGroup.Wait()
	}
	exitFn(0)
}

// MustNotError checks if error is not nil. If it is not nil it will panic.
func mustNotError(err error) {
	if err != nil {
		panic(err)
	}
}
