package godo

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mgutz/minimist"
	"github.com/mgutz/str"
	"gopkg.in/godo.v2/glob"
	"gopkg.in/godo.v2/util"
	"gopkg.in/godo.v2/watcher"
)

// TaskFunction is the signature of the function used to define a type.
// type TaskFunc func(string, ...interface{}) *Task
// type UseFunc func(string, interface{})

// A Task is an operation performed on a user's project directory.
type Task struct {
	Name         string
	description  string
	Handler      Handler
	dependencies Series
	argm         minimist.ArgMap

	// Watches are watches files. On change the task is rerun. For example `**/*.less`
	// Usually Watches and Sources are the same.
	// WatchFiles   []*FileAsset
	// WatchGlobs   []string
	// WatchRegexps []*RegexpInfo

	// computed based on dependencies
	EffectiveWatchRegexps []*glob.RegexpInfo
	EffectiveWatchGlobs   []string

	// Complete indicates whether this task has already ran. This flag is
	// ignored in watch mode.
	Complete bool
	debounce time.Duration
	RunOnce  bool

	SrcFiles   []*glob.FileAsset
	SrcGlobs   []string
	SrcRegexps []*glob.RegexpInfo

	DestFiles   []*glob.FileAsset
	DestGlobs   []string
	DestRegexps []*glob.RegexpInfo

	// used when a file event is received between debounce intervals, the file event
	// will queue itself and set this flag and force debounce to run it
	// when time has elapsed
	sync.Mutex
	ignoreEvents bool
}

// NewTask creates a new Task.
func NewTask(name string, argm minimist.ArgMap) *Task {
	runOnce := false
	if strings.HasSuffix(name, "?") {
		runOnce = true
		name = str.ChompRight(name, "?")
	}
	return &Task{Name: name, RunOnce: runOnce, dependencies: Series{}, argm: argm}
}

// Expands glob patterns.
func (task *Task) expandGlobs() {

	// runs once lazily
	if len(task.SrcFiles) > 0 {
		return
	}

	files, regexps, err := glob.Glob(task.SrcGlobs)
	if err != nil {
		util.Error(task.Name, "%v", err)
		return
	}

	task.SrcRegexps = regexps
	task.SrcFiles = files

	if len(task.DestGlobs) > 0 {
		files, regexps, err := glob.Glob(task.DestGlobs)
		if err != nil {
			util.Error(task.Name, "%v", err)
			return
		}
		task.DestRegexps = regexps
		task.DestFiles = files
	}
}

// Run runs all the dependencies of this task and when they have completed,
// runs this task.
func (task *Task) Run() error {
	if !watching && task.Complete {
		util.Debug(task.Name, "Already ran\n")
		return nil
	}
	return task.RunWithEvent(task.Name, nil)
}

// isWatchedFile determines if a FileEvent's file is a watched file
func (task *Task) isWatchedFile(path string) bool {
	filename, err := filepath.Rel(wd, path)
	if err != nil {
		return false
	}

	filename = filepath.ToSlash(filename)
	//util.Debug("task", "checking for match %s\n", filename)

	matched := false
	for _, info := range task.EffectiveWatchRegexps {
		if info.Negate {
			if matched {
				matched = !info.MatchString(filename)
				//util.Debug("task", "negated match? %s %s\n", filename, matched)
				continue
			}
		} else if info.MatchString(filename) {
			matched = true
			//util.Debug("task", "matched %s %s\n", filename, matched)
			continue
		}
	}
	return matched
}

// RunWithEvent runs this task when triggered from a watch.
// *e* FileEvent contains information about the file/directory which changed
// in watch mode.
func (task *Task) RunWithEvent(logName string, e *watcher.FileEvent) (err error) {
	if task.RunOnce && task.Complete {
		util.Debug(task.Name, "Already ran\n")
		return nil
	}

	task.expandGlobs()
	if !task.shouldRun(e) {
		util.Info(logName, "up-to-date 0ms\n")
		return nil
	}

	start := time.Now()
	if len(task.SrcGlobs) > 0 && len(task.SrcFiles) == 0 {
		util.Error("task", "\""+task.Name+"\" '%v' did not match any files\n", task.SrcGlobs)
	}

	// Run this task only if the file matches watch Regexps
	rebuilt := ""
	if e != nil {
		rebuilt = "rebuilt "
		if !task.isWatchedFile(e.Path) && len(task.SrcGlobs) > 0 {
			return nil
		}
		if verbose {
			util.Debug(logName, "%s\n", e.String())
		}
	}

	log := true
	if task.Handler != nil {
		context := Context{Task: task, Args: task.argm, FileEvent: e}
		defer func() {
			if p := recover(); p != nil {
				sp, ok := p.(*softPanic)
				if !ok {
					panic(p)
				}
				err = fmt.Errorf("%q: %s", logName, sp)
			}
		}()

		task.Handler.Handle(&context)
		if context.Error != nil {
			return fmt.Errorf("%q: %s", logName, context.Error.Error())
		}
	} else if len(task.dependencies) > 0 {
		// no need to log if just dependency
		log = false
	} else {
		util.Info(task.Name, "Ignored. Task does not have a handler or dependencies.\n")
		return nil
	}

	if log {
		if rebuilt != "" {
			util.InfoColorful(logName, "%s%vms\n", rebuilt, time.Since(start).Nanoseconds()/1e6)
		} else {
			util.Info(logName, "%s%vms\n", rebuilt, time.Since(start).Nanoseconds()/1e6)
		}
	}

	task.Complete = true

	return nil
}

// DependencyNames gets the flattened dependency names.
func (task *Task) DependencyNames() []string {
	if len(task.dependencies) == 0 {
		return nil
	}
	deps := []string{}
	for _, dep := range task.dependencies {
		switch d := dep.(type) {
		default:
			panic("dependencies can only be Serial or Parallel")
		case Series:
			deps = append(deps, d.names()...)
		case Parallel:
			deps = append(deps, d.names()...)
		case S:
			deps = append(deps, Series(d).names()...)
		case P:
			deps = append(deps, Parallel(d).names()...)
		}
	}
	return deps
}

func (task *Task) dump(buf io.Writer, indent string) {
	fmt.Fprintln(buf, indent, task.Name)
	fmt.Fprintln(buf, indent+indent, "EffectiveWatchGlobs", task.EffectiveWatchGlobs)
	fmt.Fprintln(buf, indent+indent, "SrcFiles", task.SrcFiles)
	fmt.Fprintln(buf, indent+indent, "SrcGlobs", task.SrcGlobs)

}

func (task *Task) shouldRun(e *watcher.FileEvent) bool {
	if e == nil || len(task.SrcFiles) == 0 {
		return true
	} else if !task.isWatchedFile(e.Path) {
		// fmt.Printf("received a file so it should return immediately\n")
		return false
	}

	// lazily expand globs
	task.expandGlobs()

	if len(task.SrcFiles) == 0 || len(task.DestFiles) == 0 {
		// fmt.Printf("no source files %s %#v\n", task.Name, task.SrcFiles)
		// fmt.Printf("no source files %s %#v\n", task.Name, task.DestFiles)
		return true
	}

	// TODO figure out intelligent way to cache this instead of stating
	// each time
	for _, src := range task.SrcFiles {
		// refresh stat
		src.Stat()
		for _, dest := range task.DestFiles {
			// refresh stat
			dest.Stat()
			if filepath.Base(src.Path) == "foo.txt" {
				fmt.Printf("src %s %#v\n", src.Path, src.ModTime().UnixNano())
				fmt.Printf("dest %s %#v\n", dest.Path, dest.ModTime().UnixNano())
			}
			if src.ModTime().After(dest.ModTime()) {
				return true
			}
		}
	}

	fmt.Printf("FileEvent ignored %#v\n", e)

	return false
}

func (task *Task) debounceValue() time.Duration {
	if task.debounce == 0 {
		// use default Debounce
		return Debounce
	}
	return task.debounce
}
