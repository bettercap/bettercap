package otto

import (
	"testing"
)

func TestNativeStackFrames(t *testing.T) {
	tt(t, func() {
		vm := New()

		s, err := vm.Compile("input.js", `
      function A() { ext1(); }
      function B() { ext2(); }
      A();
    `)
		if err != nil {
			panic(err)
		}

		vm.Set("ext1", func(c FunctionCall) Value {
			if _, err := c.Otto.Eval("B()"); err != nil {
				panic(err)
			}

			return UndefinedValue()
		})

		vm.Set("ext2", func(c FunctionCall) Value {
			{
				// no limit, include innermost native frames
				ctx := c.Otto.ContextSkip(-1, false)

				is(ctx.Stacktrace, []string{
					"github.com/robertkrimen/otto.TestNativeStackFrames.func1.2 (native_stack_test.go:28)",
					"B (input.js:3:22)",
					"github.com/robertkrimen/otto.TestNativeStackFrames.func1.1 (native_stack_test.go:20)",
					"A (input.js:2:22)", "input.js:4:7",
				})

				is(ctx.Callee, "github.com/robertkrimen/otto.TestNativeStackFrames.func1.2")
				is(ctx.Filename, "native_stack_test.go")
				is(ctx.Line, 28)
				is(ctx.Column, 0)
			}

			{
				// no limit, skip innermost native frames
				ctx := c.Otto.ContextSkip(-1, true)

				is(ctx.Stacktrace, []string{
					"B (input.js:3:22)",
					"github.com/robertkrimen/otto.TestNativeStackFrames.func1.1 (native_stack_test.go:20)",
					"A (input.js:2:22)", "input.js:4:7",
				})

				is(ctx.Callee, "B")
				is(ctx.Filename, "input.js")
				is(ctx.Line, 3)
				is(ctx.Column, 22)
			}

			if _, err := c.Otto.Eval("ext3()"); err != nil {
				panic(err)
			}

			return UndefinedValue()
		})

		vm.Set("ext3", func(c FunctionCall) Value {
			{
				// no limit, include innermost native frames
				ctx := c.Otto.ContextSkip(-1, false)

				is(ctx.Stacktrace, []string{
					"github.com/robertkrimen/otto.TestNativeStackFrames.func1.3 (native_stack_test.go:69)",
					"github.com/robertkrimen/otto.TestNativeStackFrames.func1.2 (native_stack_test.go:28)",
					"B (input.js:3:22)",
					"github.com/robertkrimen/otto.TestNativeStackFrames.func1.1 (native_stack_test.go:20)",
					"A (input.js:2:22)", "input.js:4:7",
				})

				is(ctx.Callee, "github.com/robertkrimen/otto.TestNativeStackFrames.func1.3")
				is(ctx.Filename, "native_stack_test.go")
				is(ctx.Line, 69)
				is(ctx.Column, 0)
			}

			{
				// no limit, skip innermost native frames
				ctx := c.Otto.ContextSkip(-1, true)

				is(ctx.Stacktrace, []string{
					"B (input.js:3:22)",
					"github.com/robertkrimen/otto.TestNativeStackFrames.func1.1 (native_stack_test.go:20)",
					"A (input.js:2:22)", "input.js:4:7",
				})

				is(ctx.Callee, "B")
				is(ctx.Filename, "input.js")
				is(ctx.Line, 3)
				is(ctx.Column, 22)
			}

			return UndefinedValue()
		})

		if _, err := vm.Run(s); err != nil {
			panic(err)
		}
	})
}
