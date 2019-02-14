package godo

import (
	"time"

	"gopkg.in/godo.v2/util"
	"github.com/mgutz/str"
)

// Dependency marks an interface as a dependency.
type Dependency interface {
	markAsDependency()
}

// Series are dependent tasks which must run in series.
type Series []interface{}

func (s Series) names() []string {
	names := []string{}
	for _, step := range s {
		switch t := step.(type) {
		case string:
			if str.SliceIndexOf(names, t) < 0 {
				names = append(names, t)
			}
		case Series:
			names = append(names, t.names()...)
		case Parallel:
			names = append(names, t.names()...)
		}

	}
	return names
}

func (s Series) markAsDependency() {}

// Parallel runs tasks in parallel
type Parallel []interface{}

func (p Parallel) names() []string {
	names := []string{}
	for _, step := range p {
		switch t := step.(type) {
		case string:
			if str.SliceIndexOf(names, t) < 0 {
				names = append(names, t)
			}
		case Series:
			names = append(names, t.names()...)
		case Parallel:
			names = append(names, t.names()...)
		}

	}
	return names
}

func (p Parallel) markAsDependency() {}

// S is alias for Series
type S []interface{}

func (s S) markAsDependency() {}

// P is alias for Parallel
type P []interface{}

func (p P) markAsDependency() {}

// Debounce is minimum milliseconds before task can run again
func (task *Task) Debounce(duration time.Duration) *Task {
	if duration > 0 {
		task.debounce = duration
	}
	return task
}

// Deps are task dependencies and must specify how to run tasks in series or in parallel.
func (task *Task) Deps(names ...interface{}) {
	for _, name := range names {
		switch dep := name.(type) {
		default:
			util.Error(task.Name, "Dependency types must be (string | P | Parallel | S | Series)")
		case string:
			task.dependencies = append(task.dependencies, dep)
		case P:
			task.dependencies = append(task.dependencies, Parallel(dep))
		case Parallel:
			task.dependencies = append(task.dependencies, dep)
		case S:
			task.dependencies = append(task.dependencies, Series(dep))
		case Series:
			task.dependencies = append(task.dependencies, dep)
		}
	}
}

// Description sets the description for the task.
func (task *Task) Description(desc string) *Task {
	if desc != "" {
		task.description = desc
	}
	return task
}

// Desc is alias for Description.
func (task *Task) Desc(desc string) *Task {
	return task.Description(desc)
}

// Dest adds target globs which are used to calculated outdated files.
// The tasks is not run unless ANY file Src are newer than ANY
// in DestN.
func (task *Task) Dest(globs ...string) *Task {
	if len(globs) > 0 {
		task.DestGlobs = globs
	}
	return task
}

// Src adds a source globs to this task. The task is
// not run unless files are outdated between Src and Dest globs.
func (task *Task) Src(globs ...string) *Task {
	if len(globs) > 0 {
		task.SrcGlobs = globs
	}
	return task
}
