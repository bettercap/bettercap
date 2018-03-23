package otto

import (
	"testing"
)

func TestError(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`
            [ Error.prototype.name, Error.prototype.message, Error.prototype.hasOwnProperty("message") ];
        `, "Error,,true")
	})
}

func TestError_instanceof(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`(new TypeError()) instanceof Error`, true)
	})
}

func TestPanicValue(t *testing.T) {
	tt(t, func() {
		test, vm := test()

		vm.Set("abc", func(call FunctionCall) Value {
			value, err := call.Otto.Run(`({ def: 3.14159 })`)
			is(err, nil)
			panic(value)
		})

		test(`
            try {
                abc();
            }
            catch (err) {
                error = err;
            }
            [ error instanceof Error, error.message, error.def ];
        `, "false,,3.14159")
	})
}

func Test_catchPanic(t *testing.T) {
	tt(t, func() {
		vm := New()

		_, err := vm.Run(`
            A syntax error that
            does not define
            var;
                abc;
        `)
		is(err, "!=", nil)

		_, err = vm.Call(`abc.def`, nil)
		is(err, "!=", nil)
	})
}

func TestErrorContext(t *testing.T) {
	tt(t, func() {
		vm := New()

		_, err := vm.Run(`
            undefined();
        `)
		{
			err := err.(*Error)
			is(err.message, "'undefined' is not a function")
			is(len(err.trace), 1)
			is(err.trace[0].location(), "<anonymous>:2:13")
		}

		_, err = vm.Run(`
            ({}).abc();
        `)
		{
			err := err.(*Error)
			is(err.message, "'abc' is not a function")
			is(len(err.trace), 1)
			is(err.trace[0].location(), "<anonymous>:2:14")
		}

		_, err = vm.Run(`
            ("abc").abc();
        `)
		{
			err := err.(*Error)
			is(err.message, "'abc' is not a function")
			is(len(err.trace), 1)
			is(err.trace[0].location(), "<anonymous>:2:14")
		}

		_, err = vm.Run(`
            var ghi = "ghi";
            ghi();
        `)
		{
			err := err.(*Error)
			is(err.message, "'ghi' is not a function")
			is(len(err.trace), 1)
			is(err.trace[0].location(), "<anonymous>:3:13")
		}

		_, err = vm.Run(`
            function def() {
                undefined();
            }
            function abc() {
                def();
            }
            abc();
        `)
		{
			err := err.(*Error)
			is(err.message, "'undefined' is not a function")
			is(len(err.trace), 3)
			is(err.trace[0].location(), "def (<anonymous>:3:17)")
			is(err.trace[1].location(), "abc (<anonymous>:6:17)")
			is(err.trace[2].location(), "<anonymous>:8:13")
		}

		_, err = vm.Run(`
            function abc() {
                xyz();
            }
            abc();
        `)
		{
			err := err.(*Error)
			is(err.message, "'xyz' is not defined")
			is(len(err.trace), 2)
			is(err.trace[0].location(), "abc (<anonymous>:3:17)")
			is(err.trace[1].location(), "<anonymous>:5:13")
		}

		_, err = vm.Run(`
            mno + 1;
        `)
		{
			err := err.(*Error)
			is(err.message, "'mno' is not defined")
			is(len(err.trace), 1)
			is(err.trace[0].location(), "<anonymous>:2:13")
		}

		_, err = vm.Run(`
            eval("xyz();");
        `)
		{
			err := err.(*Error)
			is(err.message, "'xyz' is not defined")
			is(len(err.trace), 1)
			is(err.trace[0].location(), "<anonymous>:1:1")
		}

		_, err = vm.Run(`
            xyzzy = "Nothing happens."
            eval("xyzzy();");
        `)
		{
			err := err.(*Error)
			is(err.message, "'xyzzy' is not a function")
			is(len(err.trace), 1)
			is(err.trace[0].location(), "<anonymous>:1:1")
		}

		_, err = vm.Run(`
            throw Error("xyzzy");
        `)
		{
			err := err.(*Error)
			is(err.message, "xyzzy")
			is(len(err.trace), 1)
			is(err.trace[0].location(), "<anonymous>:2:19")
		}

		_, err = vm.Run(`
            throw new Error("xyzzy");
        `)
		{
			err := err.(*Error)
			is(err.message, "xyzzy")
			is(len(err.trace), 1)
			is(err.trace[0].location(), "<anonymous>:2:23")
		}

		script1, err := vm.Compile("file1.js",
			`function A() {
				throw new Error("test");
			}

			function C() {
				var o = null;
				o.prop = 1;
			}
		`)
		is(err, nil)

		_, err = vm.Run(script1)
		is(err, nil)

		script2, err := vm.Compile("file2.js",
			`function B() {
				A()
			}
		`)
		is(err, nil)

		_, err = vm.Run(script2)
		is(err, nil)

		script3, err := vm.Compile("file3.js", "B()")
		is(err, nil)

		_, err = vm.Run(script3)
		{
			err := err.(*Error)
			is(err.message, "test")
			is(len(err.trace), 3)
			is(err.trace[0].location(), "A (file1.js:2:15)")
			is(err.trace[1].location(), "B (file2.js:2:5)")
			is(err.trace[2].location(), "file3.js:1:1")
		}

		{
			f, _ := vm.Get("B")
			_, err := f.Call(UndefinedValue())
			err1 := err.(*Error)
			is(err1.message, "test")
			is(len(err1.trace), 2)
			is(err1.trace[0].location(), "A (file1.js:2:15)")
			is(err1.trace[1].location(), "B (file2.js:2:5)")
		}

		{
			f, _ := vm.Get("C")
			_, err := f.Call(UndefinedValue())
			err1 := err.(*Error)
			is(err1.message, "Cannot access member 'prop' of null")
			is(len(err1.trace), 1)
			is(err1.trace[0].location(), "C (file1.js:7:5)")
		}

	})
}

func TestMakeCustomErrorReturn(t *testing.T) {
	tt(t, func() {
		vm := New()

		vm.Set("A", func(c FunctionCall) Value {
			return vm.MakeCustomError("CarrotError", "carrots is life, carrots is love")
		})

		s, _ := vm.Compile("test.js", `
			function B() { return A(); }
			function C() { return B(); }
			function D() { return C(); }
		`)

		if _, err := vm.Run(s); err != nil {
			panic(err)
		}

		v, err := vm.Call("D", nil)
		if err != nil {
			panic(err)
		}

		is(v.Class(), "Error")

		name, err := v.Object().Get("name")
		if err != nil {
			panic(err)
		}
		is(name.String(), "CarrotError")

		message, err := v.Object().Get("message")
		if err != nil {
			panic(err)
		}
		is(message.String(), "carrots is life, carrots is love")

		str, err := v.Object().Call("toString")
		if err != nil {
			panic(err)
		}
		is(str, "CarrotError: carrots is life, carrots is love")

		i, err := v.Export()
		if err != nil {
			panic(err)
		}
		t.Logf("%#v\n", i)
	})
}

func TestMakeCustomError(t *testing.T) {
	tt(t, func() {
		vm := New()

		vm.Set("A", func(c FunctionCall) Value {
			panic(vm.MakeCustomError("CarrotError", "carrots is life, carrots is love"))

			return UndefinedValue()
		})

		s, _ := vm.Compile("test.js", `
			function B() { A(); }
			function C() { B(); }
			function D() { C(); }
		`)

		if _, err := vm.Run(s); err != nil {
			panic(err)
		}

		_, err := vm.Call("D", nil)
		if err == nil {
			panic("error should not be nil")
		}

		is(err.Error(), "CarrotError: carrots is life, carrots is love")

		er := err.(*Error)

		is(er.name, "CarrotError")
		is(er.message, "carrots is life, carrots is love")
	})
}

func TestMakeCustomErrorFreshVM(t *testing.T) {
	tt(t, func() {
		vm := New()
		e := vm.MakeCustomError("CarrotError", "carrots is life, carrots is love")

		str, err := e.ToString()
		if err != nil {
			panic(err)
		}

		is(str, "CarrotError: carrots is life, carrots is love")
	})
}

func TestMakeTypeError(t *testing.T) {
	tt(t, func() {
		vm := New()

		vm.Set("A", func(c FunctionCall) Value {
			panic(vm.MakeTypeError("these aren't my glasses"))

			return UndefinedValue()
		})

		s, _ := vm.Compile("test.js", `
			function B() { A(); }
			function C() { B(); }
			function D() { C(); }
		`)

		if _, err := vm.Run(s); err != nil {
			panic(err)
		}

		_, err := vm.Call("D", nil)
		if err == nil {
			panic("error should not be nil")
		}

		is(err.Error(), "TypeError: these aren't my glasses")

		er := err.(*Error)

		is(er.name, "TypeError")
		is(er.message, "these aren't my glasses")
	})
}

func TestMakeRangeError(t *testing.T) {
	tt(t, func() {
		vm := New()

		vm.Set("A", func(c FunctionCall) Value {
			panic(vm.MakeRangeError("too many"))

			return UndefinedValue()
		})

		s, _ := vm.Compile("test.js", `
			function B() { A(); }
			function C() { B(); }
			function D() { C(); }
		`)

		if _, err := vm.Run(s); err != nil {
			panic(err)
		}

		_, err := vm.Call("D", nil)
		if err == nil {
			panic("error should not be nil")
		}

		is(err.Error(), "RangeError: too many")

		er := err.(*Error)

		is(er.name, "RangeError")
		is(er.message, "too many")
	})
}

func TestMakeSyntaxError(t *testing.T) {
	tt(t, func() {
		vm := New()

		vm.Set("A", func(c FunctionCall) Value {
			panic(vm.MakeSyntaxError("i think you meant \"you're\""))

			return UndefinedValue()
		})

		s, _ := vm.Compile("test.js", `
			function B() { A(); }
			function C() { B(); }
			function D() { C(); }
		`)

		if _, err := vm.Run(s); err != nil {
			panic(err)
		}

		_, err := vm.Call("D", nil)
		if err == nil {
			panic("error should not be nil")
		}

		is(err.Error(), "SyntaxError: i think you meant \"you're\"")

		er := err.(*Error)

		is(er.name, "SyntaxError")
		is(er.message, "i think you meant \"you're\"")
	})
}

func TestErrorStackProperty(t *testing.T) {
	tt(t, func() {
		vm := New()

		s, err := vm.Compile("test.js", `
			function A() { throw new TypeError('uh oh'); }
			function B() { return A(); }
			function C() { return B(); }

			var s = null;

			try { C(); } catch (e) { s = e.stack; }

			s;
		`)
		if err != nil {
			panic(err)
		}

		v, err := vm.Run(s)
		if err != nil {
			panic(err)
		}

		is(v.String(), "TypeError: uh oh\n    at A (test.js:2:29)\n    at B (test.js:3:26)\n    at C (test.js:4:26)\n    at test.js:8:10\n")
	})
}
