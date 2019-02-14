package godo

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mgutz/minimist"
	"gopkg.in/godo.v2/glob"
	"gopkg.in/godo.v2/util"
	"gopkg.in/godo.v2/watcher"
)

// softPanic is used to check for errors within a task handler.
type softPanic struct {
	// msg is the original error that caused the panic
	err error
}

func (sp *softPanic) Error() string {
	return sp.err.Error()
}

// Halt is a soft panic and stops a task.
func Halt(v interface{}) {
	if v == nil {
		panic("No reason provided")
	} else if err, ok := v.(error); ok {
		panic(&softPanic{err})
	}

	panic(&softPanic{fmt.Errorf("%v", v)})
}

// M is generic string to interface alias
type M map[string]interface{}

// Project is a container for tasks.
type Project struct {
	sync.Mutex
	Tasks       map[string]*Task
	Namespace   map[string]*Project
	lastRun     map[string]time.Time
	exitFn      func(code int)
	ns          string
	contextArgm minimist.ArgMap
	cwatchTasks map[chan bool]bool

	parent *Project
}

// NewProject creates am empty project ready for tasks.
func NewProject(tasksFunc func(*Project), exitFn func(code int), argm minimist.ArgMap) *Project {
	project := &Project{Tasks: map[string]*Task{}, lastRun: map[string]time.Time{}}
	project.Namespace = map[string]*Project{}
	project.Namespace[""] = project
	project.ns = "root"
	project.exitFn = exitFn
	project.contextArgm = argm
	project.Define(tasksFunc)
	project.cwatchTasks = map[chan bool]bool{}
	return project
}

// reset resets project state
func (project *Project) reset() {
	for _, task := range project.Tasks {
		task.Complete = false
	}
	project.lastRun = map[string]time.Time{}
}

func (project *Project) mustTask(name string) (*Project, *Task, string) {
	if name == "" {
		panic("Cannot get task for empty string")
	}

	proj := project

	// use root
	if strings.HasPrefix(name, "/") {
		name = name[1:]
		for true {
			if proj.parent != nil {
				proj = proj.parent
			} else {
				break
			}
		}
	} else {
		proj = project
	}

	taskName := "default"
	parts := strings.Split(name, ":")

	if len(parts) == 1 {
		taskName = parts[0]
	} else {
		namespace := ""

		for i := 0; i < len(parts)-1; i++ {

			if namespace != "" {
				namespace += ":"
			}
			ns := parts[i]
			namespace += ns

			proj = proj.Namespace[ns]
			if proj == nil {
				util.Panic("ERR", "Could not find project having namespace \"%s\"\n", namespace)
			}
		}
		taskName = parts[len(parts)-1]
	}

	task := proj.Tasks[taskName]
	if task == nil {
		util.Panic("ERR", `"%s" task is not defined`+"\n", name)
	}
	return proj, task, taskName
}

func (project *Project) debounce(task *Task) bool {
	if task.Name == "" {
		panic("task name should not be empty")
	}
	debounce := task.debounce
	if debounce == 0 {
		debounce = Debounce
	}

	now := time.Now()
	project.Lock()
	defer project.Unlock()

	oldRun := project.lastRun[task.Name]
	if oldRun.IsZero() {
		project.lastRun[task.Name] = now
		return false
	}

	if oldRun.Add(debounce).After(now) {
		project.lastRun[task.Name] = now
		return true
	}
	return false
}

// Run runs a task by name.
func (project *Project) Run(name string) error {
	return project.run(name, name, nil)
}

// RunWithEvent runs a task by name and adds FileEvent e to the context.
func (project *Project) runWithEvent(name string, logName string, e *watcher.FileEvent) error {
	return project.run(name, logName, e)
}

func (project *Project) runTask(depName string, parentName string, e *watcher.FileEvent) error {
	proj, _, taskName := project.mustTask(depName)

	if proj == nil {
		return fmt.Errorf("Project was not loaded for \"%s\" task", parentName)
	}
	return proj.runWithEvent(taskName, parentName+">"+depName, e)
}

func (project *Project) runParallel(steps []interface{}, parentName string, e *watcher.FileEvent) error {
	var funcs = []func() error{}
	for _, step := range steps {
		switch t := step.(type) {
		default:
			panic(parentName + ": Parallel flow can only have types: (string | Series | Parallel)")
		case string:
			funcs = append(funcs, func() error {
				return project.runTask(t, parentName, e)
			})
		case S:
			funcs = append(funcs, func() error {
				return project.runSeries(t, parentName, e)
			})
		case Series:
			funcs = append(funcs, func() error {
				return project.runSeries(t, parentName, e)
			})
		case P:
			funcs = append(funcs, func() error {
				return project.runParallel(t, parentName, e)
			})
		case Parallel:
			funcs = append(funcs, func() error {
				return project.runParallel(t, parentName, e)
			})
		}
	}
	err := GoThrottle(3, funcs...)
	return err
}

func (project *Project) runSeries(steps []interface{}, parentName string, e *watcher.FileEvent) error {
	var err error
	for _, step := range steps {
		switch t := step.(type) {
		default:
			panic(parentName + ": Series can only have types: (string | Series | Parallel)")
		case string:
			err = project.runTask(t, parentName, e)
		case S:
			err = project.runSeries(t, parentName, e)
		case Series:
			err = project.runSeries(t, parentName, e)
		case P:
			err = project.runParallel(t, parentName, e)
		case Parallel:
			err = project.runParallel(t, parentName, e)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// run runs the project, executing any tasks named on the command line.
func (project *Project) run(name string, logName string, e *watcher.FileEvent) error {
	proj, task, _ := project.mustTask(name)

	if !task.shouldRun(e) {
		return nil
	}

	// debounce needs to be separate from shouldRun, so we can enqueue
	// a file event that arrives between debounce intervals
	if proj.debounce(task) {
		if task.shouldRun(e) {
			task.Lock()
			if !task.ignoreEvents {
				task.ignoreEvents = true
				// fmt.Printf("DBG: ENQUEUE fileevent in between debounce\n")
				time.AfterFunc(task.debounceValue(), func() {
					// fmt.Printf("DBG: Running ENQUEUED\n")
					task.Lock()
					task.ignoreEvents = false
					task.Unlock()
					project.run(name, logName, e)
				})
			}
			task.Unlock()
		}

		return nil
	}

	// run dependencies first
	err := proj.runSeries(task.dependencies, name, e)
	if err != nil {
		return err
	}

	// then run the task itself
	return task.RunWithEvent(logName, e)
}

// usage returns a string for usage screen
func (project *Project) usage() string {
	tasks := "Tasks:\n"
	names := []string{}
	m := map[string]*Task{}
	for ns, proj := range project.Namespace {
		if ns != "" {
			ns += ":"
		}
		for _, task := range proj.Tasks {
			names = append(names, ns+task.Name)
			m[ns+task.Name] = task
		}
	}
	sort.Strings(names)
	longest := 0
	for _, name := range names {
		l := len(name)
		if l > longest {
			longest = l
		}
	}

	for _, name := range names {
		task := m[name]
		description := task.description
		if description == "" {
			if len(task.dependencies) > 0 {
				description = fmt.Sprintf("Runs %v %s", task.DependencyNames(), name)
			} else {
				description = "Runs " + name
			}
		}
		tasks += fmt.Sprintf("  %-"+strconv.Itoa(longest)+"s  %s\n", name, description)
	}

	return tasks
}

// Use uses another project's task within a namespace.
func (project *Project) Use(namespace string, tasksFunc func(*Project)) {
	namespace = strings.Trim(namespace, ":")
	proj := NewProject(tasksFunc, project.exitFn, project.contextArgm)
	proj.ns = project.ns + ":" + namespace
	project.Namespace[namespace] = proj
	proj.parent = project
}

// Task adds a task to the project with dependencies and handler.
func (project *Project) Task(name string, dependencies Dependency, handler func(*Context)) *Task {
	task := NewTask(name, project.contextArgm)

	if handler == nil && dependencies == nil {
		util.Panic("godo", "Task %s requires a dependency or handler\n", name)
	}

	if handler != nil {
		task.Handler = HandlerFunc(handler)
	}
	if dependencies != nil {
		task.dependencies = append(task.dependencies, dependencies)
	}

	project.Tasks[task.Name] = task
	return task
}

// Task1 adds a simple task to the project.
func (project *Project) Task1(name string, handler func(*Context)) *Task {
	task := NewTask(name, project.contextArgm)

	if handler == nil {
		util.Panic("godo", "Task %s requires a dependency or handler\n", name)
	}

	task.Handler = HandlerFunc(handler)

	project.Tasks[task.Name] = task
	return task
}

// TaskD adds a task which runs other dependencies with no handler.
func (project *Project) TaskD(name string, dependencies Dependency) *Task {
	task := NewTask(name, project.contextArgm)

	if dependencies == nil {
		util.Panic("godo", "Task %s requires a dependency or handler\n", name)
	}

	task.dependencies = append(task.dependencies, dependencies)
	project.Tasks[task.Name] = task
	return task
}

func (project *Project) watchTask(task *Task, root string, logName string, handler func(e *watcher.FileEvent)) {
	ignorePathFn := func(p string) bool {
		return watcher.DefaultIgnorePathFn(p) || !task.isWatchedFile(p)
	}

	const bufferSize = 2048
	watchr, err := watcher.NewWatcher(bufferSize)
	if err != nil {
		util.Panic("project", "%v\n", err)
	}
	watchr.IgnorePathFn = ignorePathFn
	watchr.ErrorHandler = func(err error) {
		util.Error("project", "Watcher error %v\n", err)
	}
	watchr.WatchRecursive(root)

	// this function will block forever, Ctrl+C to quit app
	abs, err := filepath.Abs(root)
	if err != nil {
		fmt.Println("Could not get absolute path", err)
		return
	}
	util.Info(logName, "watching %s\n", abs)

	// not sure why this need to be unbuffered, but it was blocking
	// on cquit <- true
	cquit := make(chan bool, 1)
	project.Lock()
	project.cwatchTasks[cquit] = true
	project.Unlock()
	watchr.Start()
forloop:
	for {
		select {
		case event := <-watchr.Event:
			if event.Path != "" {
				util.InfoColorful("godo", "%s changed\n", event.Path)
			}
			handler(event)
		case <-cquit:
			watchr.Stop()
			break forloop
		}
	}
}

// Define defines tasks
func (project *Project) Define(fn func(*Project)) {
	fn(project)
}

func calculateWatchPaths(patterns []string) []string {
	//fmt.Println("DBG:calculateWatchPaths patterns", patterns)
	paths := map[string]bool{}
	for _, pat := range patterns {
		if pat == "" {
			continue
		}
		path := glob.PatternRoot(pat)
		abs, err := filepath.Abs(path)
		if err != nil {
			fmt.Println("Error calculating watch paths", err)
		}
		paths[abs] = true
	}

	var keys []string
	for key := range paths {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	//fmt.Println("DBG:calculateWatchPaths keys", keys)

	// skip any directories that overlap each other, eg test/sub should be
	// ignored if test/ is in paths
	var skip = map[string]bool{}
	for i, dir := range keys {
		dirSlash := dir + "/"
		for _, dirj := range keys[i+1:] {
			if strings.HasPrefix(dirj, dirSlash) {
				skip[dirj] = true
			}
		}
	}

	var keep = []string{}
	for _, dir := range keys {
		if skip[dir] {
			continue
		}
		rel, err := filepath.Rel(wd, dir)
		if err != nil {
			fmt.Println("Error calculating relative path", err)
			continue
		}
		keep = append(keep, rel)
	}

	//fmt.Println("DBG:calculateWatchPaths keep", keep)
	return keep
}

// gatherWatchInfo updates globs and regexps for the task based on its dependencies
func (project *Project) gatherWatchInfo(task *Task) (globs []string, regexps []*glob.RegexpInfo) {
	globs = task.SrcGlobs
	regexps = task.SrcRegexps

	if len(task.dependencies) > 0 {
		names := task.DependencyNames()

		proj := project
		for _, depname := range names {
			var task *Task
			proj, task, _ = project.mustTask(depname)
			tglobs, tregexps := proj.gatherWatchInfo(task)
			task.EffectiveWatchRegexps = tregexps
			globs = append(globs, tglobs...)
			regexps = append(regexps, tregexps...)
		}
	}
	task.EffectiveWatchRegexps = regexps
	task.EffectiveWatchGlobs = globs
	return
}

// Watch watches the Files of a task and reruns the task on a watch event. Any
// direct dependency is also watched. Returns true if watching.
//
//
// TODO:
// 1. Only the parent task watches, but it gathers wath info from all dependencies.
//
// 2. Anything without src files always run when a dependency is triggered by a glob match.
//
//		build [generate{*.go} compile] => go file changes =>  build, generate and compile
//
// 3. Tasks with src only run if it matches a src
//
//       build [generate{*.go} css{*.scss} compile] => go file changes => build, generate and compile
//       css does not need to run since no SCSS files ran
//
// X depends on [A:txt, B]	=> txt changes	A runs, X runs without deps
// X:txt on [A, B]			=> txt changes	A, B, X runs
//
func (project *Project) Watch(names []string, isParent bool) bool {
	// fixes a bug where the first debounce prevents the task from running because
	// all tasks are run once before Watch() is called
	project.reset()

	funcs := []func(){}

	taskClosure := func(project *Project, task *Task, taskname string, logName string) func() {
		paths := calculateWatchPaths(task.EffectiveWatchGlobs)
		return func() {
			if len(paths) == 0 {
				return
			}
			for _, pth := range paths {
				go func(path string) {
					project.watchTask(task, path, logName, func(e *watcher.FileEvent) {
						err := project.run(taskname, taskname, e)
						if err != nil {
							util.Error("ERR", "%s\n", err.Error())
						}
					})
				}(pth)
			}
		}
	}

	for _, taskname := range names {
		proj, task, _ := project.mustTask(taskname)
		// updates effectiveWatchGlobs
		proj.gatherWatchInfo(task)
		if len(task.EffectiveWatchGlobs) > 0 {
			funcs = append(funcs, taskClosure(project, task, taskname, taskname))
		}
	}

	if len(funcs) > 0 {
		<-all(funcs)
		return true
	}
	return false
}

// Dumps information about the project to the console
func (project *Project) dump(buf io.Writer, prefix string, indent string) {
	fmt.Fprintln(buf, "")
	fmt.Fprintln(buf, prefix, project.ns, " =>")
	fmt.Fprintln(buf, indent, "Tasks:")
	for _, task := range project.Tasks {
		task.dump(buf, indent+indent)
	}

	for key, proj := range project.Namespace {
		if key == "" {
			continue
		}
		proj.dump(buf, prefix, indent)
	}
}

func (project *Project) quit(isParent bool) {
	for ns, proj := range project.Namespace {
		if ns != "" {
			proj.quit(false)
		}
	}
	// kill all watchTasks
	for cquit := range project.cwatchTasks {
		cquit <- true
	}
	if isParent {
		runnerWaitGroup.Stop()
		for _, process := range Processes {
			if process != nil {
				process.Kill()
			}
		}
	}
	//fmt.Printf("DBG: QUITTED\n")
}

// Exit quits the project.
func (project *Project) Exit(code int) {
	project.quit(true)
}

// all runs the functions in fns concurrently.
func all(fns []func()) (done <-chan bool) {
	var wg sync.WaitGroup
	wg.Add(len(fns))

	ch := make(chan bool, 1)
	for _, fn := range fns {
		go func(f func()) {
			f()
			wg.Done()
		}(fn)
	}
	go func() {
		wg.Wait()
		doneSig(ch, true)
	}()
	return ch
}

func doneSig(ch chan bool, val bool) {
	ch <- val
	close(ch)
}
