package log

import (
	"github.com/evilsocket/bettercap-ng/session"
)

func Debug(format string, args ...interface{}) {
	session.I.Events.Log(session.DEBUG, format, args)
}

func Info(format string, args ...interface{}) {
	session.I.Events.Log(session.INFO, format, args)
}

func Warning(format string, args ...interface{}) {
	session.I.Events.Log(session.WARNING, format, args)
}

func Error(format string, args ...interface{}) {
	session.I.Events.Log(session.ERROR, format, args)
}

func Fatal(format string, args ...interface{}) {
	session.I.Events.Log(session.FATAL, format, args)
}
