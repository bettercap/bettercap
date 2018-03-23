package otto

import (
	"testing"
)

// this is its own file because the tests in it rely on the line numbers of
// some of the functions defined here. putting it in with the rest of the
// tests would probably be annoying.

func TestErrorContextNative(t *testing.T) {
	tt(t, func() {
		vm := New()

		vm.Set("N", func(c FunctionCall) Value {
			v, err := c.Argument(0).Call(NullValue())
			if err != nil {
				panic(err)
			}
			return v
		})

		s, _ := vm.Compile("test.js", `
			function F() { throw new Error('wow'); }
			function G() { return N(F); }
		`)

		vm.Run(s)

		f1, _ := vm.Get("G")
		_, err := f1.Call(NullValue())
		err1 := err.(*Error)
		is(err1.message, "wow")
		is(len(err1.trace), 3)
		is(err1.trace[0].location(), "F (test.js:2:29)")
		is(err1.trace[1].location(), "github.com/robertkrimen/otto.TestErrorContextNative.func1.1 (error_native_test.go:15)")
		is(err1.trace[2].location(), "G (test.js:3:26)")
	})
}
