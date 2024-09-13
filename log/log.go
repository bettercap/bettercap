package log

import (
	"github.com/evilsocket/islazy/log"
)

type logFunction func(level log.Verbosity, format string, args ...interface{})

var Logger = (logFunction)(nil)

func Debug(format string, args ...interface{}) {
	if Logger != nil {
		Logger(log.DEBUG, format, args...)
	}
}

func Info(format string, args ...interface{}) {
	if Logger != nil {
		Logger(log.INFO, format, args...)
	}
}

func Warning(format string, args ...interface{}) {
	if Logger != nil {
		Logger(log.WARNING, format, args...)
	}
}

func Error(format string, args ...interface{}) {
	if Logger != nil {
		Logger(log.ERROR, format, args...)
	}
}

func Fatal(format string, args ...interface{}) {
	if Logger != nil {
		Logger(log.FATAL, format, args...)
	}
}
