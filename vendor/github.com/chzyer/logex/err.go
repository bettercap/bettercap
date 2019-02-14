package logex

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"runtime"
	"strconv"
	"strings"
)

func Define(info string) *traceError {
	return &traceError{
		error: errors.New(info),
	}
}

func NewError(info ...interface{}) *traceError {
	return TraceEx(1, errors.New(fmt.Sprint(info...)))
}

func NewErrorf(format string, info ...interface{}) *traceError {
	return TraceEx(1, fmt.Errorf(format, info...))
}

func EqualAny(e error, es []error) bool {
	for i := 0; i < len(es); i++ {
		if Equal(e, es[i]) {
			return true
		}
	}
	return false
}

func Equal(e1, e2 error) bool {
	if e, ok := e1.(*traceError); ok {
		e1 = e.error
	}
	if e, ok := e2.(*traceError); ok {
		e2 = e.error
	}
	return e1 == e2
}

type traceError struct {
	error
	format []interface{}
	stack  []string
	code   *int
}

func (t *traceError) SetCode(code int) *traceError {
	if t.stack == nil {
		t = TraceEx(1, t)
	}
	t.code = &code
	return t
}

func (t *traceError) GetCode() int {
	if t.code == nil {
		return 500
	}
	return *t.code
}

func (t *traceError) Error() string {
	if t == nil {
		return "<nil>"
	}
	if t.format == nil {
		if t.error == nil {
			panic(t.stack)
		}
		return t.error.Error()
	}
	return fmt.Sprintf(t.error.Error(), t.format...)
}

func (t *traceError) Trace(info ...interface{}) *traceError {
	return TraceEx(1, t, info...)
}

func (t *traceError) Follow(err error) *traceError {
	if t == nil {
		return nil
	}
	if te, ok := err.(*traceError); ok {
		if len(te.stack) > 0 {
			te.stack[len(te.stack)-1] += ":" + err.Error()
		}
		t.stack = append(te.stack, t.stack...)
	}
	return t
}

func (t *traceError) Format(obj ...interface{}) *traceError {
	if t.stack == nil {
		t = TraceEx(1, t)
	}
	t.format = obj
	return t
}

func (t *traceError) StackError() string {
	if t == nil {
		return t.Error()
	}
	if len(t.stack) == 0 {
		return t.Error()
	}
	return fmt.Sprintf("[%s] %s", strings.Join(t.stack, ";"), t.Error())
}

func Tracef(err error, obj ...interface{}) *traceError {
	e := TraceEx(1, err).Format(obj...)
	return e
}

// set runtime info to error
func TraceError(err error, info ...interface{}) *traceError {
	return TraceEx(1, err, info...)
}

func Trace(err error, info ...interface{}) error {
	if err == nil {
		return nil
	}
	return TraceEx(1, err, info...)
}

func joinInterface(info []interface{}, ch string) string {
	ret := bytes.NewBuffer(make([]byte, 0, 512))
	for idx, o := range info {
		if idx > 0 {
			ret.WriteString(ch)
		}
		ret.WriteString(fmt.Sprint(o))
	}
	return ret.String()
}

func TraceEx(depth int, err error, info ...interface{}) *traceError {
	if err == nil {
		return nil
	}
	pc, _, line, _ := runtime.Caller(1 + depth)
	name := runtime.FuncForPC(pc).Name()
	name = path.Base(name)
	stack := name + ":" + strconv.Itoa(line)
	if len(info) > 0 {
		stack += "(" + joinInterface(info, ",") + ")"
	}
	if te, ok := err.(*traceError); ok {
		if te.stack == nil { // define
			return &traceError{
				error: te.error,
				stack: []string{stack},
			}
		}
		te.stack = append(te.stack, stack)
		return te
	}
	return &traceError{err, nil, []string{stack}, nil}
}
