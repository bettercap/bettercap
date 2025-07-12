package log

import (
	"testing"

	"github.com/evilsocket/islazy/log"
)

var called bool
var calledLevel log.Verbosity
var calledFormat string
var calledArgs []interface{}

func mockLogger(level log.Verbosity, format string, args ...interface{}) {
	called = true
	calledLevel = level
	calledFormat = format
	calledArgs = args
}

func reset() {
	called = false
	calledLevel = log.DEBUG
	calledFormat = ""
	calledArgs = nil
}

func TestLoggerNil(t *testing.T) {
	reset()
	Logger = nil

	Debug("test")
	if called {
		t.Error("Debug should not call if Logger is nil")
	}

	Info("test")
	if called {
		t.Error("Info should not call if Logger is nil")
	}

	Warning("test")
	if called {
		t.Error("Warning should not call if Logger is nil")
	}

	Error("test")
	if called {
		t.Error("Error should not call if Logger is nil")
	}

	Fatal("test")
	if called {
		t.Error("Fatal should not call if Logger is nil")
	}
}

func TestDebug(t *testing.T) {
	reset()
	Logger = mockLogger

	Debug("test %d", 42)
	if !called || calledLevel != log.DEBUG || calledFormat != "test %d" || len(calledArgs) != 1 || calledArgs[0] != 42 {
		t.Errorf("Debug not called correctly: level=%v format=%s args=%v", calledLevel, calledFormat, calledArgs)
	}
}

func TestInfo(t *testing.T) {
	reset()
	Logger = mockLogger

	Info("test %s", "info")
	if !called || calledLevel != log.INFO || calledFormat != "test %s" || len(calledArgs) != 1 || calledArgs[0] != "info" {
		t.Errorf("Info not called correctly: level=%v format=%s args=%v", calledLevel, calledFormat, calledArgs)
	}
}

func TestWarning(t *testing.T) {
	reset()
	Logger = mockLogger

	Warning("test %f", 3.14)
	if !called || calledLevel != log.WARNING || calledFormat != "test %f" || len(calledArgs) != 1 || calledArgs[0] != 3.14 {
		t.Errorf("Warning not called correctly: level=%v format=%s args=%v", calledLevel, calledFormat, calledArgs)
	}
}

func TestError(t *testing.T) {
	reset()
	Logger = mockLogger

	Error("test error")
	if !called || calledLevel != log.ERROR || calledFormat != "test error" || len(calledArgs) != 0 {
		t.Errorf("Error not called correctly: level=%v format=%s args=%v", calledLevel, calledFormat, calledArgs)
	}
}

func TestFatal(t *testing.T) {
	reset()
	Logger = mockLogger

	Fatal("test fatal")
	if !called || calledLevel != log.FATAL || calledFormat != "test fatal" || len(calledArgs) != 0 {
		t.Errorf("Fatal not called correctly: level=%v format=%s args=%v", calledLevel, calledFormat, calledArgs)
	}
}
