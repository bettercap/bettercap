package log

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testBuf bytes.Buffer

var testInternalLog Logger

func init() {
	testInternalLog = NewLogger3(&testBuf, "__logxi", NewTextFormatter("__logxi"))
	testInternalLog.SetLevel(LevelError)
}

func TestUnknownLevel(t *testing.T) {
	testResetEnv()
	os.Setenv("LOGXI", "*=oy")
	processEnv()
	buffer := testBuf.String()
	assert.Contains(t, buffer, "Unknown level", "should error on unknown level")
}
