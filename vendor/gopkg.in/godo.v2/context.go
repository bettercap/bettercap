package godo

import (
	"github.com/mgutz/minimist"
	"gopkg.in/godo.v2/util"
	"gopkg.in/godo.v2/watcher"
)

func logVerbose(msg string, format string, args ...interface{}) {
	if !verbose {
		return
	}
	util.Debug(msg, format, args...)
}

// Context is the data passed to a task.
type Context struct {
	// Task is the currently running task.
	Task *Task

	// FileEvent is an event from the watcher with change details.
	FileEvent *watcher.FileEvent

	// Task command line arguments
	Args minimist.ArgMap

	Error error
}

// AnyFile returns either a non-DELETe FileEvent file or the WatchGlob patterns which
// can be used by goa.Load()
func (context *Context) AnyFile() []string {
	if context.FileEvent != nil && context.FileEvent.Event != watcher.DELETED {
		return []string{context.FileEvent.Path}
	}
	return context.Task.SrcGlobs
}

// Run runs a command
func (context *Context) Run(cmd string, options ...map[string]interface{}) {
	if context.Error != nil {
		logVerbose(context.Task.Name, "Context is in error. Skipping: %s\n", cmd)
		return
	}
	_, err := Run(cmd, options...)
	if err != nil {
		context.Error = err
	}
}

// Bash runs a bash shell.
func (context *Context) Bash(cmd string, options ...map[string]interface{}) {
	if context.Error != nil {
		logVerbose(context.Task.Name, "Context is in error. Skipping: %s\n", cmd)
		return
	}
	_, err := Bash(cmd, options...)
	if err != nil {
		context.Error = err
	}
}

// Start run aysnchronously.
func (context *Context) Start(cmd string, options ...map[string]interface{}) {
	if context.Error != nil {
		logVerbose(context.Task.Name, "Context is in error. Skipping: %s\n", cmd)
		return
	}

	err := startEx(context, cmd, options)
	if err != nil {
		context.Error = err
	}
}

// BashOutput executes a bash script and returns the output
func (context *Context) BashOutput(script string, options ...map[string]interface{}) string {
	if len(options) == 0 {
		options = append(options, M{"$out": CaptureBoth})
	} else {
		options[0]["$out"] = CaptureBoth
	}
	s, err := Bash(script, options...)
	if err != nil {
		context.Error = err
		return ""
	}
	return s
}

// RunOutput runs a command and returns output.
func (context *Context) RunOutput(commandstr string, options ...map[string]interface{}) string {
	if len(options) == 0 {
		options = append(options, M{"$out": CaptureBoth})
	} else {
		options[0]["$out"] = CaptureBoth
	}
	s, err := Run(commandstr, options...)
	if err != nil {
		context.Error = err
		return ""
	}
	return s
}

// Check halts the task if err is not nil.
//
// Do this
//		Check(err, "Some error occured")
//
// Instead of
//
//		if err != nil {
//			Halt(err)
//		}
func (context *Context) Check(err error, msg string) {
	if err != nil {
		if msg == "" {
			Halt(err)
			return
		}
		Halt(msg + ": " + err.Error())
	}
}
