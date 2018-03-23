package otto

import (
	"testing"
)

const (
	testSourcemapCodeOriginal  = "function functionA(argA, argB) {\n  functionB(argA, argB);\n}\n\nfunction functionB(argA, argB) {\n  functionExternal(argA, argB);\n}"
	testSourcemapCodeMangled   = "function functionA(argA,argB){functionB(argA,argB)}function functionB(argA,argB){functionExternal(argA,argB)}"
	testSourcemapContent       = `{"version":3,"sources":["hello.js"],"names":["functionA","argA","argB","functionB","functionExternal"],"mappings":"AAAA,QAASA,WAAUC,KAAMC,MACvBC,UAAUF,KAAMC,MAGlB,QAASC,WAAUF,KAAMC,MACvBE,iBAAiBH,KAAMC"}`
	testSourcemapInline        = "function functionA(argA,argB){functionB(argA,argB)}function functionB(argA,argB){functionExternal(argA,argB)}\n//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbImhlbGxvLmpzIl0sIm5hbWVzIjpbImZ1bmN0aW9uQSIsImFyZ0EiLCJhcmdCIiwiZnVuY3Rpb25CIiwiZnVuY3Rpb25FeHRlcm5hbCJdLCJtYXBwaW5ncyI6IkFBQUEsUUFBU0EsV0FBVUMsS0FBTUMsTUFDdkJDLFVBQVVGLEtBQU1DLE1BR2xCLFFBQVNDLFdBQVVGLEtBQU1DLE1BQ3ZCRSxpQkFBaUJILEtBQU1DIn0="
	testSourcemapOriginalStack = "ReferenceError: 'functionExternal' is not defined\n    at functionB (hello.js:6:3)\n    at functionA (hello.js:2:3)\n    at <anonymous>:1:1\n"
	testSourcemapMangledStack  = "ReferenceError: 'functionExternal' is not defined\n    at functionB (hello.js:1:82)\n    at functionA (hello.js:1:31)\n    at <anonymous>:1:1\n"
	testSourcemapMappedStack   = "ReferenceError: 'functionExternal' is not defined\n    at functionB (hello.js:6:2)\n    at functionA (hello.js:2:2)\n    at <anonymous>:1:1\n"
)

func TestSourceMapOriginalWithNoSourcemap(t *testing.T) {
	tt(t, func() {
		vm := New()

		s, err := vm.Compile("hello.js", testSourcemapCodeOriginal)
		if err != nil {
			panic(err)
		}

		if _, err := vm.Run(s); err != nil {
			panic(err)
		}

		_, err = vm.Run(`functionA()`)
		if err == nil {
			panic("error should not be nil")
		}

		is(err.(*Error).String(), testSourcemapOriginalStack)
	})
}

func TestSourceMapMangledWithNoSourcemap(t *testing.T) {
	tt(t, func() {
		vm := New()

		s, err := vm.Compile("hello.js", testSourcemapCodeMangled)
		if err != nil {
			panic(err)
		}

		if _, err := vm.Run(s); err != nil {
			panic(err)
		}

		_, err = vm.Run(`functionA()`)
		if err == nil {
			panic("error should not be nil")
		}

		is(err.(*Error).String(), testSourcemapMangledStack)
	})
}

func TestSourceMapMangledWithSourcemap(t *testing.T) {
	tt(t, func() {
		vm := New()

		s, err := vm.CompileWithSourceMap("hello.js", testSourcemapCodeMangled, testSourcemapContent)
		if err != nil {
			panic(err)
		}

		if _, err := vm.Run(s); err != nil {
			panic(err)
		}

		_, err = vm.Run(`functionA()`)
		if err == nil {
			panic("error should not be nil")
		}

		is(err.(*Error).String(), testSourcemapMappedStack)
	})
}

func TestSourceMapMangledWithInlineSourcemap(t *testing.T) {
	tt(t, func() {
		vm := New()

		s, err := vm.CompileWithSourceMap("hello.js", testSourcemapInline, nil)
		if err != nil {
			panic(err)
		}

		if _, err := vm.Run(s); err != nil {
			panic(err)
		}

		_, err = vm.Run(`functionA()`)
		if err == nil {
			panic("error should not be nil")
		}

		is(err.(*Error).String(), testSourcemapMappedStack)
	})
}

func TestSourceMapContextPosition(t *testing.T) {
	tt(t, func() {
		vm := New()

		s, err := vm.CompileWithSourceMap("hello.js", testSourcemapCodeMangled, testSourcemapContent)
		if err != nil {
			panic(err)
		}

		if _, err := vm.Run(s); err != nil {
			panic(err)
		}

		vm.Set("functionExternal", func(c FunctionCall) Value {
			ctx := c.Otto.Context()

			is(ctx.Filename, "hello.js")
			is(ctx.Line, 6)
			is(ctx.Column, 2)

			return UndefinedValue()
		})

		if _, err := vm.Run(`functionA()`); err != nil {
			panic(err)
		}
	})
}

func TestSourceMapContextStacktrace(t *testing.T) {
	tt(t, func() {
		vm := New()

		s, err := vm.CompileWithSourceMap("hello.js", testSourcemapCodeMangled, testSourcemapContent)
		if err != nil {
			panic(err)
		}

		if _, err := vm.Run(s); err != nil {
			panic(err)
		}

		vm.Set("functionExternal", func(c FunctionCall) Value {
			ctx := c.Otto.Context()

			is(ctx.Stacktrace, []string{
				"functionB (hello.js:6:2)",
				"functionA (hello.js:2:2)",
				"<anonymous>:1:1",
			})

			return UndefinedValue()
		})

		if _, err := vm.Run(`functionA()`); err != nil {
			panic(err)
		}
	})
}
