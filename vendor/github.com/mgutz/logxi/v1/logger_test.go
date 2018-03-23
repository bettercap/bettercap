package log

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func processEnv() {
	ProcessEnv(readFromEnviron())
}

func testResetEnv() {
	disableColors = false
	testBuf.Reset()
	os.Clearenv()
	processEnv()
	InternalLog = testInternalLog
}

func TestEnvLOGXI(t *testing.T) {
	assert := assert.New(t)

	os.Setenv("LOGXI", "")
	processEnv()
	assert.Equal(LevelWarn, logxiNameLevelMap["*"], "Unset LOGXI defaults to *:WRN with TTY")

	// default all to ERR
	os.Setenv("LOGXI", "*=ERR")
	processEnv()
	level := getLogLevel("mylog")
	assert.Equal(LevelError, level)
	level = getLogLevel("mylog2")
	assert.Equal(LevelError, level)

	// unrecognized defaults to LevelDebug on TTY
	os.Setenv("LOGXI", "mylog=badlevel")
	processEnv()
	level = getLogLevel("mylog")
	assert.Equal(LevelWarn, level)

	// wildcard should not override exact match
	os.Setenv("LOGXI", "*=WRN,mylog=ERR,other=OFF")
	processEnv()
	level = getLogLevel("mylog")
	assert.Equal(LevelError, level)
	level = getLogLevel("other")
	assert.Equal(LevelOff, level)

	// wildcard pattern should match
	os.Setenv("LOGXI", "*log=ERR")
	processEnv()
	level = getLogLevel("mylog")
	assert.Equal(LevelError, level, "wildcat prefix should match")

	os.Setenv("LOGXI", "myx*=ERR")
	processEnv()
	level = getLogLevel("mylog")
	assert.Equal(LevelError, level, "no match should return LevelError")

	os.Setenv("LOGXI", "myl*,-foo")
	processEnv()
	level = getLogLevel("mylog")
	assert.Equal(LevelAll, level)
	level = getLogLevel("foo")
	assert.Equal(LevelOff, level)
}

func TestEnvLOGXI_FORMAT(t *testing.T) {
	assert := assert.New(t)
	oldIsTerminal := isTerminal

	os.Setenv("LOGXI_FORMAT", "")
	setDefaults(true)
	processEnv()
	assert.Equal(FormatHappy, logxiFormat, "terminal defaults to FormatHappy")
	setDefaults(false)
	processEnv()
	assert.Equal(FormatJSON, logxiFormat, "non terminal defaults to FormatJSON")

	os.Setenv("LOGXI_FORMAT", "JSON")
	processEnv()
	assert.Equal(FormatJSON, logxiFormat)

	os.Setenv("LOGXI_FORMAT", "json")
	setDefaults(true)
	processEnv()
	assert.Equal(FormatHappy, logxiFormat, "Mismatches defaults to FormatHappy")
	setDefaults(false)
	processEnv()
	assert.Equal(FormatJSON, logxiFormat, "Mismatches defaults to FormatJSON non terminal")

	isTerminal = oldIsTerminal
	setDefaults(isTerminal)
}

func TestEnvLOGXI_COLORS(t *testing.T) {
	oldIsTerminal := isTerminal

	os.Setenv("LOGXI_COLORS", "*=off")
	setDefaults(true)
	processEnv()

	var buf bytes.Buffer
	l := NewLogger3(&buf, "telc", NewHappyDevFormatter("logxi-colors"))
	l.SetLevel(LevelDebug)
	l.Info("info")

	r := regexp.MustCompile(`^\d{2}:\d{2}:\d{2}\.\d{6} INF logxi-colors info`)
	assert.True(t, r.Match(buf.Bytes()))

	setDefaults(true)

	isTerminal = oldIsTerminal
	setDefaults(isTerminal)
}

func TestComplexKeys(t *testing.T) {
	testResetEnv()
	var buf bytes.Buffer
	l := NewLogger(&buf, "bench")
	assert.Panics(t, func() {
		l.Error("complex", "foo\n", 1)
	})

	assert.Panics(t, func() {
		l.Error("complex", "foo\"s", 1)
	})

	l.Error("apos is ok", "foo's", 1)
}

func TestJSON(t *testing.T) {
	testResetEnv()
	var buf bytes.Buffer
	l := NewLogger3(&buf, "bench", NewJSONFormatter("bench"))
	l.SetLevel(LevelDebug)
	l.Error("hello", "foo", "bar")

	var obj map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &obj)
	assert.NoError(t, err)
	assert.Equal(t, "bar", obj["foo"].(string))
	assert.Equal(t, "hello", obj[KeyMap.Message].(string))
}

func TestJSONImbalanced(t *testing.T) {
	testResetEnv()
	var buf bytes.Buffer
	l := NewLogger3(&buf, "bench", NewJSONFormatter("bench"))
	l.SetLevel(LevelDebug)
	l.Error("hello", "foo", "bar", "bah")

	var obj map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &obj)
	assert.NoError(t, err)
	assert.Exactly(t, []interface{}{"foo", "bar", "bah"}, obj[warnImbalancedKey])
	assert.Equal(t, "hello", obj[KeyMap.Message].(string))
}

func TestJSONNoArgs(t *testing.T) {
	testResetEnv()
	var buf bytes.Buffer
	l := NewLogger3(&buf, "bench", NewJSONFormatter("bench"))
	l.SetLevel(LevelDebug)
	l.Error("hello")

	var obj map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &obj)
	assert.NoError(t, err)
	assert.Equal(t, "hello", obj[KeyMap.Message].(string))
}

func TestJSONNested(t *testing.T) {
	testResetEnv()
	var buf bytes.Buffer
	l := NewLogger3(&buf, "bench", NewJSONFormatter("bench"))
	l.SetLevel(LevelDebug)
	l.Error("hello", "obj", map[string]string{"fruit": "apple"})

	var obj map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &obj)
	assert.NoError(t, err)
	assert.Equal(t, "hello", obj[KeyMap.Message].(string))
	o := obj["obj"]
	assert.Equal(t, "apple", o.(map[string]interface{})["fruit"].(string))
}

func TestJSONEscapeSequences(t *testing.T) {
	testResetEnv()
	var buf bytes.Buffer
	l := NewLogger3(&buf, "bench", NewJSONFormatter("bench"))
	l.SetLevel(LevelDebug)
	esc := "I said, \"a's \\ \\\b\f\n\r\t\x1a\"你好'; DELETE FROM people"

	var obj map[string]interface{}
	// test as message
	l.Error(esc)
	err := json.Unmarshal(buf.Bytes(), &obj)
	assert.NoError(t, err)
	assert.Equal(t, esc, obj[KeyMap.Message].(string))

	// test as key
	buf.Reset()
	key := "你好"
	l.Error("as key", key, "esc")
	err = json.Unmarshal(buf.Bytes(), &obj)
	assert.NoError(t, err)
	assert.Equal(t, "as key", obj[KeyMap.Message].(string))
	assert.Equal(t, "esc", obj[key].(string))
}

func TestKeyNotString(t *testing.T) {
	testResetEnv()
	var buf bytes.Buffer
	l := NewLogger3(&buf, "badkey", NewHappyDevFormatter("badkey"))
	l.SetLevel(LevelDebug)
	l.Debug("foo", 1)
	assert.Panics(t, func() {
		l.Debug("reserved key", "_t", "trying to use time")
	})
}

func TestWarningErrorContext(t *testing.T) {
	testResetEnv()
	var buf bytes.Buffer
	l := NewLogger3(&buf, "wrnerr", NewHappyDevFormatter("wrnerr"))
	l.Warn("no keys")
	l.Warn("has eys", "key1", 2)
	l.Error("no keys")
	l.Error("has keys", "key1", 2)
}

func TestLevels(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger3(&buf, "bench", NewJSONFormatter("bench"))

	l.SetLevel(LevelFatal)
	assert.False(t, l.IsWarn())
	assert.False(t, l.IsInfo())
	assert.False(t, l.IsTrace())
	assert.False(t, l.IsDebug())

	l.SetLevel(LevelError)
	assert.False(t, l.IsWarn())

	l.SetLevel(LevelWarn)
	assert.True(t, l.IsWarn())
	assert.False(t, l.IsDebug())

	l.SetLevel(LevelInfo)
	assert.True(t, l.IsInfo())
	assert.True(t, l.IsWarn())
	assert.False(t, l.IsDebug())

	l.SetLevel(LevelDebug)
	assert.True(t, l.IsDebug())
	assert.True(t, l.IsInfo())
	assert.False(t, l.IsTrace())

	l.SetLevel(LevelTrace)
	assert.True(t, l.IsTrace())
	assert.True(t, l.IsDebug())
}

func TestAllowSingleParam(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger3(&buf, "wrnerr", NewTextFormatter("wrnerr"))
	l.SetLevel(LevelDebug)
	l.Info("info", 1)
	assert.True(t, strings.HasSuffix(buf.String(), singleArgKey+": 1\n"))

	buf.Reset()
	l = NewLogger3(&buf, "wrnerr", NewHappyDevFormatter("wrnerr"))
	l.SetLevel(LevelDebug)
	l.Info("info", 1)
	assert.True(t, strings.HasSuffix(buf.String(), "_: \x1b[0m1\n"))

	var obj map[string]interface{}
	buf.Reset()
	l = NewLogger3(&buf, "wrnerr", NewJSONFormatter("wrnerr"))
	l.SetLevel(LevelDebug)
	l.Info("info", 1)
	err := json.Unmarshal(buf.Bytes(), &obj)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), obj["_"])
}

func TestErrorOnWarn(t *testing.T) {
	testResetEnv()
	// os.Setenv("LOGXI_FORMAT", "context=2")
	// processEnv()
	var buf bytes.Buffer
	l := NewLogger3(&buf, "wrnerr", NewHappyDevFormatter("wrnerr"))
	l.SetLevel(LevelWarn)

	ErrorDummy := errors.New("dummy error")

	err := l.Warn("warn with error", "err", ErrorDummy)
	assert.Error(t, err)
	assert.Equal(t, "dummy error", err.Error())
	err = l.Warn("warn with no error", "one", 1)
	assert.NoError(t, err)
	//l.Error("error with err", "err", ErrorDummy)
}

type CheckStringer struct {
	s string
}

func (cs CheckStringer) String() string {
	return "bbb"
}

func TestStringer(t *testing.T) {
	f := CheckStringer{s: "aaa"}

	var buf bytes.Buffer
	l := NewLogger3(&buf, "cs1", NewTextFormatter("stringer-text"))
	l.SetLevel(LevelDebug)
	l.Info("info", "f", f)
	assert.True(t, strings.Contains(buf.String(), "bbb"))

	buf.Reset()
	l = NewLogger3(&buf, "cs2", NewHappyDevFormatter("stringer-happy"))
	l.SetLevel(LevelDebug)
	l.Info("info", "f", f)
	assert.True(t, strings.Contains(buf.String(), "bbb"))

	var obj map[string]interface{}
	buf.Reset()
	l = NewLogger3(&buf, "cs3", NewJSONFormatter("stringer-json"))
	l.SetLevel(LevelDebug)
	l.Info("info", "f", f)
	err := json.Unmarshal(buf.Bytes(), &obj)
	assert.NoError(t, err)
	assert.Equal(t, "bbb", obj["f"])
}

// When log functions cast pointers to interface{}.
// Say p is a pointer set to nil:
//
//     interface{}(p) == nil  // this is false
//
// Casting it to interface{} makes it trickier to test whether its nil.
func TestStringerNullPointers(t *testing.T) {
	var f *CheckStringer
	var buf bytes.Buffer
	l := NewLogger3(&buf, "cs1", NewJSONFormatter("stringer-json"))
	l.SetLevel(LevelDebug)
	l.Info("info", "f", f)
	assert.Contains(t, buf.String(), "null")
}