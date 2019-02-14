package test

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/chzyer/logex"
)

var (
	mainRoot           = ""
	RootPath           = os.TempDir()
	ErrNotExcept       = logex.Define("result not expect")
	ErrNotEqual        = logex.Define("result not equals")
	ErrRequireNotEqual = logex.Define("result require not equals")
	StrNotSuchFile     = "no such file or directory"
)

func init() {
	println("tmpdir:", RootPath)
}

type testException struct {
	depth int
	info  string
}

func getMainRoot() string {
	if mainRoot != "" {
		return mainRoot
	}

	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	for len(cwd) > 1 {
		_, err := os.Stat(filepath.Join(cwd, ".git"))
		if err == nil {
			mainRoot = cwd + string([]rune{filepath.Separator})
			break
		}
		cwd = filepath.Dir(cwd)
	}
	return mainRoot
}

func Skip() {
	panic(nil)
}

type Failer interface {
	FailNow()
}

func New(t Failer) {
	err := recover()
	if err == nil {
		return
	}
	te, ok := err.(*testException)
	if !ok {
		panic(err)
	}

	_, file, line, _ := runtime.Caller(5 + te.depth)
	if strings.HasPrefix(file, getMainRoot()) {
		file = file[len(getMainRoot()):]
	}
	println(fmt.Sprintf("%s:%d: %s", file, line, te.info))
	t.FailNow()
}

func getErr(def error, e []error) error {
	if len(e) == 0 {
		return def
	}
	return e[0]
}

func ReadAt(r io.ReaderAt, b []byte, at int64) {
	n, err := r.ReadAt(b, at)
	if err != nil {
		Panic(0, fmt.Errorf("ReadAt error: %v", err))
	}
	if n != len(b) {
		Panic(0, fmt.Errorf("ReadAt short read: %v, want: %v", n, len(b)))
	}
}

func ReadAndCheck(r io.Reader, b []byte) {
	buf := make([]byte, len(b))
	Read(r, buf)
	equalBytes(1, buf, b)
}

func Read(r io.Reader, b []byte) {
	n, err := r.Read(b)
	if err != nil && !logex.Equal(err, io.EOF) {
		Panic(0, fmt.Errorf("Read error: %v", err))
	}
	if n != len(b) {
		Panic(0, fmt.Errorf("Read: %v, want: %v", n, len(b)))
	}
}

func ReadStringAt(r io.ReaderAt, off int64, s string) {
	buf := make([]byte, len(s))
	n, err := r.ReadAt(buf, off)
	buf = buf[:n]
	if err != nil {
		Panic(0, fmt.Errorf("ReadStringAt: %v", err))
	}
	if string(buf) != s {
		Panic(0, fmt.Errorf(
			"ReadStringAt not match: %v, got: %v",
			strconv.Quote(s),
			strconv.Quote(string(buf)),
		))
	}
}

func ReadString(r io.Reader, s string) {
	buf := make([]byte, len(s))
	n, err := r.Read(buf)
	if err != nil && !logex.Equal(err, io.EOF) {
		Panic(0, fmt.Errorf("ReadString: %v, got: %v", strconv.Quote(s), err))
	}
	if n != len(buf) {
		Panic(0, fmt.Errorf("ReadString: %v, got: %v", strconv.Quote(s), n))
	}
	if string(buf) != s {
		Panic(0, fmt.Errorf(
			"ReadString not match: %v, got: %v",
			strconv.Quote(s),
			strconv.Quote(string(buf)),
		))
	}
}

func WriteAt(w io.WriterAt, b []byte, at int64) {
	n, err := w.WriteAt(b, at)
	if err != nil {
		Panic(0, err)
	}
	if n != len(b) {
		Panic(0, "short write")
	}
}

func Write(w io.Writer, b []byte) {
	n, err := w.Write(b)
	if err != nil {
		Panic(0, err)
	}
	if n != len(b) {
		Panic(0, "short write")
	}
}

func WriteString(w io.Writer, s string) {
	n, err := w.Write([]byte(s))
	if err != nil {
		Panic(0, err)
	}
	if n != len(s) {
		Panic(0, "short write")
	}
}

func Equals(o ...interface{}) {
	if len(o)%2 != 0 {
		Panic(0, "invalid Equals arguments")
	}
	for i := 0; i < len(o); i += 2 {
		equal(1, o[i], o[i+1], nil)
	}
}

func NotEqual(a, b interface{}, e ...error) {
	notEqual(1, a, b, e)
}

func toInt(a interface{}) (int64, bool) {
	switch n := a.(type) {
	case int:
		return int64(n), true
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return int64(n), true
	case uintptr:
		return int64(n), true
	default:
		return -1, false
	}
}

func MarkLine() {
	r := strings.Repeat("-", 20)
	println(r)
}

var globalMarkInfo string

func Mark(obj ...interface{}) {
	globalMarkInfo = fmt.Sprint(obj...)
}

func EqualBytes(got, want []byte) {
	equalBytes(0, got, want)
}

func equalBytes(n int, got, want []byte) {
	a := got
	b := want
	size := 16
	if len(a) != len(b) {
		Panic(n, fmt.Sprintf("equal bytes, %v != %v", len(a), len(b)))
	}
	if bytes.Equal(a, b) {
		return
	}

	for off := 0; off < len(a); off += size {
		end := off + size
		if end > len(a) {
			end = len(a)
		}
		if !bytes.Equal(a[off:end], b[off:end]) {
			Panic(n, fmt.Sprintf(
				"equal [%v]byte in [%v, %v]:\n\tgot: %v\n\twant: %v",
				len(a),
				off, off+size,
				a[off:end], b[off:end],
			))
		}
	}
}

func Equal(a, b interface{}, e ...error) {
	if ai, ok := toInt(a); ok {
		if bi, ok := toInt(b); ok {
			equal(1, ai, bi, e)
			return
		}
	}
	equal(1, a, b, e)
}

func CheckError(e error, s string) {
	if e == nil {
		Panic(0, ErrNotExcept)
	}
	if !strings.Contains(e.Error(), s) {
		Panic(0, fmt.Errorf(
			"want: %s, got %s",
			strconv.Quote(s),
			strconv.Quote(e.Error()),
		))
	}
}

func formatMax(o interface{}, max int) string {
	aStr := fmt.Sprint(o)
	if len(aStr) > max {
		aStr = aStr[:max] + " ..."
	}
	return aStr
}

func notEqual(d int, a, b interface{}, e []error) {
	_, oka := a.(error)
	_, okb := b.(error)
	if oka && okb {
		if logex.Equal(a.(error), b.(error)) {
			Panic(d, fmt.Sprintf("%v: %v",
				getErr(ErrRequireNotEqual, e),
				a,
			))
		}
		return
	}
	if reflect.DeepEqual(a, b) {
		Panic(d, fmt.Sprintf("%v: (%v, %v)",
			getErr(ErrRequireNotEqual, e),
			formatMax(a, 100),
			formatMax(b, 100),
		))
	}
}

func equal(d int, a, b interface{}, e []error) {
	_, oka := a.(error)
	_, okb := b.(error)
	if oka && okb {
		if !logex.Equal(a.(error), b.(error)) {
			Panic(d, fmt.Sprintf("%v: (%v, %v)",
				getErr(ErrNotEqual, e),
				formatMax(a, 100), formatMax(b, 100),
			))
		}
		return
	}
	if !reflect.DeepEqual(a, b) {
		Panic(d, fmt.Sprintf("%v: (%+v, %+v)", getErr(ErrNotEqual, e), a, b))
	}
}

func Should(b bool, e ...error) {
	if !b {
		Panic(0, getErr(ErrNotExcept, e))
	}
}

func NotNil(obj interface{}) {
	if obj == nil {
		Panic(0, "should not nil")
	}
}

func False(obj bool) {
	if obj {
		Panic(0, "should false")
	}
}

func True(obj bool) {
	if !obj {
		Panic(0, "should true")
	}
}

func Nil(obj interface{}) {
	if obj != nil {
		// double check, incase different type with nil value
		if !reflect.ValueOf(obj).IsNil() {
			str := fmt.Sprint(obj)
			if err, ok := obj.(error); ok {
				str = logex.DecodeError(err)
			}
			Panic(0, fmt.Sprintf("should nil: %v", str))
		}
	}
}

func Panic(depth int, obj interface{}) {
	t := &testException{
		depth: depth,
	}
	if err, ok := obj.(error); ok {
		t.info = logex.DecodeError(err)
	} else {
		t.info = fmt.Sprint(obj)
	}
	if globalMarkInfo != "" {
		t.info = "[info:" + globalMarkInfo + "] " + t.info
	}
	panic(t)
}

func CleanTmp() {
	os.RemoveAll(root(2))
}

func TmpFile() (*os.File, error) {
	dir := root(2)
	if err := os.MkdirAll(dir, 0744); err != nil {
		return nil, err
	}
	return ioutil.TempFile(dir, "")
}

func Root() string {
	p := root(2)
	os.RemoveAll(root(2))
	return p
}

func root(n int) string {
	pc, _, _, _ := runtime.Caller(n)
	name := runtime.FuncForPC(pc).Name()
	if idx := strings.LastIndex(name, "."); idx > 0 {
		name = name[:idx] + "/" + name[idx+1:]
	}

	root := os.Getenv("TEST_ROOT")
	if root == "" {
		root = RootPath
	}
	return filepath.Join(root, name)
}

func RandBytes(n int) []byte {
	buf := make([]byte, n)
	rand.Read(buf)
	return buf
}

func SeqBytes(n int) []byte {
	buf := make([]byte, n)
	for idx := range buf {
		buf[idx] = byte(idx)
	}
	return buf
}
