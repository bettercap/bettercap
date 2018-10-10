package log

import (
	"github.com/bettercap/bettercap/session"

	ll "github.com/evilsocket/islazy/log"
)

func Debug(format string, args ...interface{}) {
	session.I.Events.Log(ll.DEBUG, format, args...)
}

func Info(format string, args ...interface{}) {
	session.I.Events.Log(ll.INFO, format, args...)
}

func Warning(format string, args ...interface{}) {
	session.I.Events.Log(ll.WARNING, format, args...)
}

func Error(format string, args ...interface{}) {
	session.I.Events.Log(ll.ERROR, format, args...)
}

func Fatal(format string, args ...interface{}) {
	session.I.Events.Log(ll.FATAL, format, args...)
}
