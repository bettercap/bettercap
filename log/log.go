package log

import (
	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/session"
)

func Debug(format string, args ...interface{}) {
	session.I.Events.Log(core.DEBUG, format, args...)
}

func Info(format string, args ...interface{}) {
	session.I.Events.Log(core.INFO, format, args...)
}

func Warning(format string, args ...interface{}) {
	session.I.Events.Log(core.WARNING, format, args...)
}

func Error(format string, args ...interface{}) {
	session.I.Events.Log(core.ERROR, format, args...)
}

func Fatal(format string, args ...interface{}) {
	session.I.Events.Log(core.FATAL, format, args...)
}
