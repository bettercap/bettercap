package logex

import (
	"encoding/json"
	"fmt"
	"io"
	goLog "log"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
)

var DebugLevel = 1

type Logger struct {
	depth  int
	reqid  string
	Logger *goLog.Logger
}

func NewLogger(l int) *Logger {
	return &Logger{l, "", goLogStd}
}

func NewLoggerEx(w io.Writer) *Logger {
	return &Logger{0, "", NewGoLog(w)}
}

func NewGoLog(w io.Writer) *goLog.Logger {
	return goLog.New(w, "", goLog.LstdFlags)
}

var goLogStd = goLog.New(os.Stderr, "", goLog.LstdFlags)
var std = NewLogger(1)
var ShowCode = true
var (
	Println    = std.Println
	Infof      = std.Infof
	Info       = std.Info
	Debug      = std.Debug
	Debugf     = std.Debugf
	Error      = std.Error
	Errorf     = std.Errorf
	Warn       = std.Warn
	PrintStack = std.PrintStack
	Stack      = std.Stack
	Panic      = std.Panic
	Fatal      = std.Fatal
	Struct     = std.Struct
	Pretty     = std.Pretty
	Todo       = std.Todo
)

func SetStd(l *Logger) {
	std = l
	Println = std.Println
	Infof = std.Infof
	Info = std.Info
	Debug = std.Debug
	Error = std.Error
	Warn = std.Warn
	PrintStack = std.PrintStack
	Stack = std.Stack
	Panic = std.Panic
	Fatal = std.Fatal
	Struct = std.Struct
	Pretty = std.Pretty
	Todo = std.Todo
}

var (
	INFO   = "[INFO] "
	ERROR  = "[ERROR] "
	PANIC  = "[PANIC] "
	DEBUG  = "[DEBUG] "
	WARN   = "[WARN] "
	FATAL  = "[FATAL] "
	STRUCT = "[STRUCT] "
	PRETTY = "[PRETTY] "
	TODO   = "[TODO] "
)

func color(col, s string) string {
	if col == "" {
		return s
	}
	return "\x1b[0;" + col + "m" + s + "\x1b[0m"
}

func init() {
	if os.Getenv("DEBUG") != "" {
		DebugLevel = 0
		ERROR = color("32", ERROR)
	}
}

func DownLevel(i int) Logger {
	return std.DownLevel(i - 1)
}

// decide to show which level's stack
func (l Logger) DownLevel(i int) Logger {
	return Logger{l.depth + i, l.reqid, l.Logger}
}

// output objects to json format
func (l Logger) Pretty(os ...interface{}) {
	content := ""
	for i := range os {
		if ret, err := json.MarshalIndent(os[i], "", "\t"); err == nil {
			content += string(ret) + "\n"
		}
	}
	l.Output(2, PRETTY+content)
}

// just print
func (l Logger) Print(o ...interface{}) {
	l.Output(2, sprint(o))
}

// just print by format
func (l Logger) Printf(layout string, o ...interface{}) {
	l.Output(2, sprintf(layout, o))
}

// just println
func (l Logger) Println(o ...interface{}) {
	l.Output(2, " "+sprint(o))
}

func (l Logger) Info(o ...interface{}) {
	l.Output(2, INFO+sprint(o))
}
func (l Logger) Infof(f string, o ...interface{}) {
	l.Output(2, INFO+sprintf(f, o))
}

func (l Logger) Debug(o ...interface{}) {
	if DebugLevel > 0 {
		return
	}
	l.Output(2, DEBUG+sprint(o))
}

func (l Logger) Debugf(f string, o ...interface{}) {
	if DebugLevel > 0 {
		return
	}
	l.Output(2, DEBUG+sprintf(f, o))
}

func (l Logger) Todo(o ...interface{}) {
	l.Output(2, TODO+sprint(o))
}

func (l Logger) Error(o ...interface{}) {
	l.Output(2, ERROR+sprint(o))
}

func (l Logger) Errorf(f string, o ...interface{}) {
	l.Output(2, ERROR+sprintf(f, o))
}

func (l Logger) Warn(o ...interface{}) {
	l.Output(2, WARN+sprint(o))
}
func (l Logger) Warnf(f string, o ...interface{}) {
	l.Output(2, WARN+sprintf(f, o))
}

func (l Logger) Panic(o ...interface{}) {
	l.Output(2, PANIC+sprint(o))
	panic(o)
}
func (l Logger) Panicf(f string, o ...interface{}) {
	info := sprintf(f, o)
	l.Output(2, PANIC+info)
	panic(info)
}

func (l Logger) Fatal(o ...interface{}) {
	l.Output(2, FATAL+sprint(o))
	os.Exit(1)
}

func (l Logger) Fatalf(f string, o ...interface{}) {
	l.Output(2, FATAL+sprintf(f, o))
	os.Exit(1)
}

func (l Logger) Struct(o ...interface{}) {
	items := make([]interface{}, 0, len(o)*2)
	for _, item := range o {
		items = append(items, item, item)
	}
	layout := strings.Repeat(", %T(%+v)", len(o))
	if len(layout) > 0 {
		layout = layout[2:]
	}
	l.Output(2, STRUCT+sprintf(layout, items))
}

func (l Logger) PrintStack() {
	Info(string(l.Stack()))
}

func (l Logger) Stack() []byte {
	a := make([]byte, 1024*1024)
	n := runtime.Stack(a, true)
	return a[:n]
}

func (l Logger) Output(calldepth int, s string) error {
	calldepth += l.depth + 1
	if l.Logger == nil {
		l.Logger = goLogStd
	}
	return l.Logger.Output(calldepth, l.makePrefix(calldepth)+s)
}

func (l Logger) makePrefix(calldepth int) string {
	if !ShowCode {
		return ""
	}
	pc, f, line, _ := runtime.Caller(calldepth)
	name := runtime.FuncForPC(pc).Name()
	name = path.Base(name) // only use package.funcname
	f = path.Base(f)       // only use filename

	tags := make([]string, 0, 3)

	pos := name + ":" + f + ":" + strconv.Itoa(line)
	tags = append(tags, pos)
	if l.reqid != "" {
		tags = append(tags, l.reqid)
	}
	return "[" + strings.Join(tags, "][") + "]"
}

func sprint(o []interface{}) string {
	decodeTraceError(o)
	return joinInterface(o, " ")
}
func sprintf(f string, o []interface{}) string {
	decodeTraceError(o)
	return fmt.Sprintf(f, o...)
}

func DecodeError(e error) string {
	if e == nil {
		return ""
	}
	if e1, ok := e.(*traceError); ok {
		return e1.StackError()
	}
	return e.Error()
}

func decodeTraceError(o []interface{}) {
	if !ShowCode {
		return
	}
	for idx, obj := range o {
		if te, ok := obj.(*traceError); ok {
			o[idx] = te.StackError()
		}
	}
}
