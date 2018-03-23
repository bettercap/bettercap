package otto

import (
	"testing"
)

// this is its own file because the tests in it rely on the line numbers of
// some of the functions defined here. putting it in with the rest of the
// tests would probably be annoying.

func TestFunction_stack(t *testing.T) {
	tt(t, func() {
		vm := New()

		s, _ := vm.Compile("fake.js", `function X(fn1, fn2, fn3) { fn1(fn2, fn3); }`)
		vm.Run(s)

		expected := []_frame{
			{native: true, nativeFile: "function_stack_test.go", nativeLine: 30, offset: 0, callee: "github.com/robertkrimen/otto.TestFunction_stack.func1.2"},
			{native: true, nativeFile: "function_stack_test.go", nativeLine: 25, offset: 0, callee: "github.com/robertkrimen/otto.TestFunction_stack.func1.1"},
			{native: false, nativeFile: "", nativeLine: 0, offset: 29, callee: "X", file: s.program.file},
			{native: false, nativeFile: "", nativeLine: 0, offset: 29, callee: "X", file: s.program.file},
		}

		vm.Set("A", func(c FunctionCall) Value {
			c.Argument(0).Call(UndefinedValue())
			return UndefinedValue()
		})

		vm.Set("B", func(c FunctionCall) Value {
			depth := 0
			for scope := c.Otto.runtime.scope; scope != nil; scope = scope.outer {
				// these properties are tested explicitly so that we don't test `.fn`,
				// which will differ from run to run
				is(scope.frame.native, expected[depth].native)
				is(scope.frame.nativeFile, expected[depth].nativeFile)
				is(scope.frame.nativeLine, expected[depth].nativeLine)
				is(scope.frame.offset, expected[depth].offset)
				is(scope.frame.callee, expected[depth].callee)
				is(scope.frame.file, expected[depth].file)
				depth++
			}

			is(depth, 4)

			return UndefinedValue()
		})

		x, _ := vm.Get("X")
		a, _ := vm.Get("A")
		b, _ := vm.Get("B")

		x.Call(UndefinedValue(), x, a, b)
	})
}
