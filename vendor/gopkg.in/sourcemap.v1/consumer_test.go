package sourcemap_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"gopkg.in/sourcemap.v1"
)

const jqSourceMapURL = "http://code.jquery.com/jquery-2.0.3.min.map"

var jqSourceMapBytes []byte

func init() {
	resp, err := http.Get(jqSourceMapURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	jqSourceMapBytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
}

type sourceMapTest struct {
	genLine      int
	genCol       int
	wantedSource string
	wantedName   string
	wantedLine   int
	wantedCol    int
}

func (test *sourceMapTest) String() string {
	return fmt.Sprintf("line=%d col=%d in file=%s", test.genLine, test.genCol, test.wantedSource)
}

func (test *sourceMapTest) assert(t *testing.T, smap *sourcemap.Consumer) {
	source, name, line, col, ok := smap.Source(test.genLine, test.genCol)
	if !ok {
		if test.wantedSource == "" &&
			test.wantedName == "" &&
			test.wantedLine == 0 &&
			test.wantedCol == 0 {
			return
		}
		t.Fatalf("Source not found for %s", test)
	}
	if source != test.wantedSource {
		t.Fatalf("file: got %q, wanted %q (%s)", source, test.wantedSource, test)
	}
	if name != test.wantedName {
		t.Fatalf("func: got %q, wanted %q (%s)", name, test.wantedName, test)
	}
	if line != test.wantedLine {
		t.Fatalf("line: got %d, wanted %d (%s)", line, test.wantedLine, test)
	}
	if col != test.wantedCol {
		t.Fatalf("column: got %d, wanted %d (%s)", col, test.wantedCol, test)
	}
}

func TestSourceMap(t *testing.T) {
	smap, err := sourcemap.Parse("", []byte(sourceMapJSON))
	if err != nil {
		t.Fatal(err)
	}

	tests := []*sourceMapTest{
		{1, 1, "/the/root/one.js", "", 1, 1},
		{1, 5, "/the/root/one.js", "", 1, 5},
		{1, 9, "/the/root/one.js", "", 1, 11},
		{1, 18, "/the/root/one.js", "bar", 1, 21},
		{1, 21, "/the/root/one.js", "", 2, 3},
		{1, 28, "/the/root/one.js", "baz", 2, 10},
		{1, 32, "/the/root/one.js", "bar", 2, 14},

		{2, 1, "/the/root/two.js", "", 1, 1},
		{2, 5, "/the/root/two.js", "", 1, 5},
		{2, 9, "/the/root/two.js", "", 1, 11},
		{2, 18, "/the/root/two.js", "n", 1, 21},
		{2, 21, "/the/root/two.js", "", 2, 3},
		{2, 28, "/the/root/two.js", "n", 2, 10},

		// Fuzzy match.
		{1, 20, "/the/root/one.js", "bar", 1, 21},
		{1, 30, "/the/root/one.js", "baz", 2, 10},
		{2, 12, "/the/root/two.js", "", 1, 11},
	}
	for _, test := range tests {
		test.assert(t, smap)
	}

	_, _, _, _, ok := smap.Source(3, 0)
	if ok {
		t.Fatal("source must not exist")
	}
}

func TestSourceRootURL(t *testing.T) {
	jsonStr := sourceMapJSON
	jsonStr = strings.Replace(jsonStr, "/the/root", "http://the/root", 1)
	jsonStr = strings.Replace(jsonStr, "one.js", "../one.js", 1)

	smap, err := sourcemap.Parse("", []byte(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	tests := []*sourceMapTest{
		{1, 1, "http://the/one.js", "", 1, 1},
		{2, 1, "http://the/root/two.js", "", 1, 1},
	}
	for _, test := range tests {
		test.assert(t, smap)
	}
}

func TestEmptySourceRootURL(t *testing.T) {
	jsonStr := sourceMapJSON
	jsonStr = strings.Replace(jsonStr, "/the/root", "", 1)
	jsonStr = strings.Replace(jsonStr, "one.js", "../one.js", 1)

	smap, err := sourcemap.Parse("http://the/root/app.min.map", []byte(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	tests := []*sourceMapTest{
		{1, 1, "http://the/one.js", "", 1, 1},
		{2, 1, "http://the/root/two.js", "", 1, 1},
	}
	for _, test := range tests {
		test.assert(t, smap)
	}
}

func TestAbsSourceURL(t *testing.T) {
	jsonStr := sourceMapJSON
	jsonStr = strings.Replace(jsonStr, "/the/root", "", 1)
	jsonStr = strings.Replace(jsonStr, "one.js", "http://the/root/one.js", 1)
	jsonStr = strings.Replace(jsonStr, "two.js", "/another/root/two.js", 1)

	testAbsSourceURL(t, "", jsonStr)
	testAbsSourceURL(t, "http://path/to/map", jsonStr)
}

func testAbsSourceURL(t *testing.T, mapURL, jsonStr string) {
	smap, err := sourcemap.Parse(mapURL, []byte(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	tests := []*sourceMapTest{
		{1, 1, "http://the/root/one.js", "", 1, 1},
		{2, 1, "/another/root/two.js", "", 1, 1},
	}
	for _, test := range tests {
		test.assert(t, smap)
	}
}

func TestJQuerySourceMap(t *testing.T) {
	smap, err := sourcemap.Parse(jqSourceMapURL, jqSourceMapBytes)
	if err != nil {
		t.Fatal(err)
	}

	tests := []*sourceMapTest{
		{1, 1, "", "", 0, 0},
		{4, 0, "", "", 0, 0},
		{4, 1, "http://code.jquery.com/jquery-2.0.3.js", "", 14, 0},
		{4, 10, "http://code.jquery.com/jquery-2.0.3.js", "window", 14, 11},
		{5, 6789, "http://code.jquery.com/jquery-2.0.3.js", "apply", 4360, 27},
		{5, 10006, "http://code.jquery.com/jquery-2.0.3.js", "apply", 4676, 8},
		{4, 553, "http://code.jquery.com/jquery-2.0.3.js", "ready", 93, 9},
		{999999, 0, "", "", 0, 0},
	}
	for _, test := range tests {
		test.assert(t, smap)
	}
}

// This is a test mapping which maps functions from two different files
// (one.js and two.js) to a minified generated source.
//
// Here is one.js:
//
//     ONE.foo = function (bar) {
//       return baz(bar);
//     };
//
// Here is two.js:
//
//     TWO.inc = function (n) {
//       return n + 1;
//     };
//
// And here is the generated code (min.js):
//
//     ONE.foo=function(a){return baz(a);};
//     TWO.inc=function(a){return a+1;};

var genCode = `exports.testGeneratedCode = "ONE.foo=function(a){return baz(a);};
TWO.inc=function(a){return a+1;};`

var sourceMapJSON = `{
  "version": 3,
  "file": "min.js",
  "names": ["bar", "baz", "n"],
  "sources": ["one.js", "two.js"],
  "sourceRoot": "/the/root",
  "mappings": "CAAC,IAAI,IAAM,SAAUA,GAClB,OAAOC,IAAID;CCDb,IAAI,IAAM,SAAUE,GAClB,OAAOA"
}`
