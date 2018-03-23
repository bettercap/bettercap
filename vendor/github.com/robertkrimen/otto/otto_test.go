package otto

import (
	"bytes"
	"io"
	"testing"

	"github.com/robertkrimen/otto/parser"
)

func TestOtto(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test("xyzzy = 2", 2)

		test("xyzzy + 2", 4)

		test("xyzzy += 16", 18)

		test("xyzzy", 18)

		test(`
            (function(){
                return 1
            })()
        `, 1)

		test(`
            (function(){
                return 1
            }).call(this)
        `, 1)

		test(`
            (function(){
                var result
                (function(){
                    result = -1
                })()
                return result
            })()
        `, -1)

		test(`
            var abc = 1
            abc || (abc = -1)
            abc
        `, 1)

		test(`
            var abc = (function(){ 1 === 1 })();
            abc;
        `, "undefined")
	})
}

func TestFunction__(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`
            function abc() {
                return 1;
            };
            abc();
        `, 1)
	})
}

func TestIf(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`
            abc = undefined;
            def = undefined;
            if (true) abc = 1
            else abc = 2;
            if (false) {
                def = 3;
            }
            else def = 4;

            [ abc, def ];
        `, "1,4")

		test(`
            if (1) {
                abc = 1;
            }
            else {
                abc = 0;
            }
            abc;
        `, 1)

		test(`
            if (0) {
                abc = 1;
            }
            else {
                abc = 0;
            }
            abc;
        `, 0)

		test(`
            abc = 0;
            if (0) {
                abc = 1;
            }
            abc;
        `, 0)

		test(`
            abc = 0;
            if (abc) {
                abc = 1;
            }
            abc;
        `, 0)
	})
}

func TestSequence(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`
            1, 2, 3;
        `, 3)
	})
}

func TestCall(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`
            Math.pow(3, 2);
        `, 9)
	})
}

func TestRunFunctionWithSetArguments(t *testing.T) {
	tt(t, func() {
		vm := New()
		vm.Run(`var sillyFunction = function(record){record.silly = true; record.answer *= -1};`)
		record := map[string]interface{}{"foo": "bar", "answer": 42}
		// Set performs a conversion that allows the map to be addressed as a Javascript object
		vm.Set("argument", record)
		_, err := vm.Run("sillyFunction(argument)")

		is(err, nil)
		is(record["answer"].(float64), -42)
		is(record["silly"].(bool), true)
	})
}

func TestRunFunctionWithArgumentsPassedToCall(t *testing.T) {
	tt(t, func() {
		vm := New()
		vm.Run(`var sillyFunction = function(record){record.silly = true; record.answer *= -1};`)
		record := map[string]interface{}{"foo": "bar", "answer": 42}
		_, err := vm.Call("sillyFunction", nil, record)

		is(err, nil)
		is(record["answer"].(float64), -42)
		is(record["silly"].(bool), true)
	})
}

func TestMember(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`
            abc = [ 0, 1, 2 ];
            def = {
                "abc": 0,
                "def": 1,
                "ghi": 2,
            };
            [ abc[2], def.abc, abc[1], def.def ];
        `, "2,0,1,1")
	})
}

func Test_this(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`
            typeof this;
        `, "object")
	})
}

func TestWhile(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`
            limit = 4
            abc = 0
            while (limit) {
                abc = abc + 1
                limit = limit - 1
            }
            abc;
        `, 4)
	})
}

func TestSwitch_break(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`
            var abc = true;
            var ghi = "Xyzzy";
            while (abc) {
                switch ('def') {
                case 'def':
                    break;
                }
                ghi = "Nothing happens.";
                abc = false;
            }
            ghi;
        `, "Nothing happens.")

		test(`
            var abc = true;
            var ghi = "Xyzzy";
            WHILE:
            while (abc) {
                switch ('def') {
                case 'def':
                    break WHILE;
                }
                ghi = "Nothing happens."
                abc = false
            }
            ghi;
        `, "Xyzzy")

		test(`
            var ghi = "Xyzzy";
            FOR:
            for (;;) {
                switch ('def') {
                case 'def':
                    break FOR;
                    ghi = "";
                }
                ghi = "Nothing happens.";
            }
            ghi;
        `, "Xyzzy")

		test(`
            var ghi = "Xyzzy";
            FOR:
            for (var jkl in {}) {
                switch ('def') {
                case 'def':
                    break FOR;
                    ghi = "Something happens.";
                }
                ghi = "Nothing happens.";
            }
            ghi;
        `, "Xyzzy")

		test(`
            var ghi = "Xyzzy";
            function jkl() {
                switch ('def') {
                case 'def':
                    break;
                    ghi = "";
                }
                ghi = "Nothing happens.";
            }
            while (abc) {
                jkl();
                abc = false;
                ghi = "Something happens.";
            }
            ghi;
        `, "Something happens.")
	})
}

func TestTryFinally(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`
            var abc;
            try {
                abc = 1;
            }
            finally {
                abc = 2;
            }
            abc;
        `, 2)

		test(`
            var abc = false, def = 0;
            do {
                def += 1;
                if (def > 100) {
                    break;
                }
                try {
                    continue;
                }
                finally {
                    abc = true;
                }
            }
            while(!abc && def < 10)
            def;
        `, 1)

		test(`
            var abc = false, def = 0, ghi = 0;
            do {
                def += 1;
                if (def > 100) {
                    break;
                }
                try {
                    throw 0;
                }
                catch (jkl) {
                    continue;
                }
                finally {
                    abc = true;
                    ghi = 11;
                }
                ghi -= 1;
            }
            while(!abc && def < 10)
            ghi;
        `, 11)

		test(`
            var abc = 0, def = 0;
            do {
                try {
                    abc += 1;
                    throw "ghi";
                }
                finally {
                    def = 1;
                    continue;
                }
                def -= 1;
            }
            while (abc < 2)
            [ abc, def ];
        `, "2,1")
	})
}

func TestTryCatch(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`
            var abc = 1;
            try {
                throw 4;
                abc = -1;
            }
            catch (xyzzy) {
                abc += xyzzy + 1;
            }
            abc;
        `, 6)

		test(`
            abc = 1;
            var def;
            try {
                try {
                    throw 4;
                    abc = -1;
                }
                catch (xyzzy) {
                    abc += xyzzy + 1;
                    throw 64;
                }
            }
            catch (xyzzy) {
                def = xyzzy;
                abc = -2;
            }
            [ def, abc ];
        `, "64,-2")
	})
}

func TestWith(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`
            var def;
            with({ abc: 9 }) {
                def = abc;
            }
            def;
        `, 9)

		test(`
            var def;
            with({ abc: function(){
                return 11;
            } }) {
                def = abc();
            }
            def;
        `, 11)
	})
}

func TestSwitch(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`
            var abc = 0;
            switch (0) {
            default:
                abc += 1;
            case 1:
                abc += 2;
            case 2:
                abc += 4;
            case 3:
                abc += 8;
            }
            abc;
        `, 15)

		test(`
            abc = 0;
            switch (3) {
            default:
                abc += 1;
            case 1:
                abc += 2;
            case 2:
                abc += 4;
            case 3:
                abc += 8;
            }
            abc;
        `, 8)

		test(`
            abc = 0;
            switch (60) {
            case 1:
                abc += 2;
            case 2:
                abc += 4;
            case 3:
                abc += 8;
            }
            abc;
        `, 0)
	})
}

func TestForIn(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`
            var abc;
            for (property in { a: 1 }) {
                abc = property;
            }
            abc;
        `, "a")

		test(`
            var ghi;
            for (property in new String("xyzzy")) {
                ghi = property;
            }
            ghi;
        `, "4")
	})
}

func TestFor(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`
            var abc = 7;
            for (i = 0; i < 3; i += 1) {
                abc += 1;
            }
            abc;
        `, 10)

		test(`
            abc = 7;
            for (i = 0; i < 3; i += 1) {
                abc += 1;
                if (i == 1) {
                    break;
                }
            }
            abc;
        `, 9)

		test(`
            abc = 7;
            for (i = 0; i < 3; i += 1) {
                if (i == 2) {
                    continue;
                }
                abc += 1;
            }
            abc;
        `, 9)

		test(`
            abc = 0;
            for (;;) {
                abc += 1;
                if (abc == 3)
                    break;
            }
            abc;
        `, 3)

		test(`
            for (abc = 0; ;) {
                abc += 1;
                if (abc == 3)
                    break;
            }
            abc;
        `, 3)

		test(`
            for (abc = 0; ; abc += 1) {
                abc += 1;
                if (abc == 3)
                    break;
            }
            abc;
        `, 3)
	})
}

func TestLabelled(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		// TODO Add emergency break

		test(`
            xyzzy: for (var abc = 0; abc <= 0; abc++) {
                for (var def = 0; def <= 1; def++) {
                    if (def === 0) {
                        continue xyzzy;
                    } else {
                    }
                }
            }
        `)

		test(`
            abc = 0
            def:
            while (true) {
                while (true) {
                    abc = abc + 1
                    if (abc > 11) {
                        break def;
                    }
                }
            }
            abc;
        `, 12)

		test(`
            abc = 0
            def:
            do {
                do {
                    abc = abc + 1
                    if (abc > 11) {
                        break def;
                    }
                } while (true)
            } while (true)
            abc;
        `, 12)
	})
}

func TestConditional(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`
            [ true ? false : true, true ? 1 : 0, false ? 3.14159 : "abc" ];
        `, "false,1,abc")
	})
}

func TestArrayLiteral(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`
            [ 1, , 3.14159 ];
        `, "1,,3.14159")
	})
}

func TestAssignment(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`
            var abc = 1;
            abc;
        `, 1)

		test(`
            abc += 2;
            abc;
        `, 3)
	})
}

func TestBinaryOperation(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`0 == 1`, false)
		test(`1 == "1"`, true)
		test(`0 === 1`, false)
		test(`1 === "1"`, false)
		test(`"1" === "1"`, true)
	})
}

func Test_typeof(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`typeof abc`, "undefined")
		test(`typeof abc === 'undefined'`, true)
		test(`typeof {}`, "object")
		test(`typeof null`, "object")
	})
}

func Test_PrimitiveValueObjectValue(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		Number11 := test(`new Number(11)`)
		is(Number11.float64(), 11)
	})
}

func Test_eval(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		// FIXME terst, Is this correct?
		test(`
            var abc = 1;
        `, "undefined")

		test(`
            eval("abc += 1");
        `, 2)

		test(`
            (function(){
                var abc = 11;
                eval("abc += 1");
                return abc;
            })();
        `, 12)
		test(`abc`, 2)

		test(`
            (function(){
                try {
                    eval("var prop = \\u2029;");
                    return false;
                } catch (abc) {
                    return [ abc instanceof SyntaxError, abc.toString() ];
                }
            })();
        `, "true,SyntaxError: Unexpected token ILLEGAL")

		test(`
            function abc(){
                this.THIS = eval("this");
            }
            var def = new abc();
            def === def.THIS;
        `, true)
	})
}

func Test_evalDirectIndirect(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		// (function () {return this;}()).abc = "global";
		test(`
            var abc = "global";
            (function(){
                try {
                    var _eval = eval;
                    var abc = "function";
                    return [
                        _eval("\'global\' === abc"),  // eval (Indirect)
                        eval("\'function\' === abc"), // eval (Direct)
                    ];
                } finally {
                    delete this.abc;
                }
            })();
        `, "true,true")
	})
}

func TestError_URIError(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`new URIError() instanceof URIError`, true)

		test(`
            var abc
            try {
                decodeURI("http://example.com/ _^#%")
            }
            catch (def) {
                abc = def instanceof URIError
            }
            abc
        `, true)
	})
}

func TestTo(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		{
			value, _ := test(`"11"`).ToFloat()
			is(value, float64(11))
		}

		{
			value, _ := test(`"11"`).ToInteger()
			is(value, int64(11))

			value, _ = test(`1.1`).ToInteger()
			is(value, int64(1))
		}
	})
}

func TestShouldError(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`raise:
            xyzzy
                throw new TypeError("Nothing happens.")
        `, "ReferenceError: 'xyzzy' is not defined")
	})
}

func TestAPI(t *testing.T) {
	tt(t, func() {
		test, vm := test()

		test(`
            String.prototype.xyzzy = function(){
                return this.length + 11 + (arguments[0] || 0)
            }
            abc = new String("xyzzy")
            def = "Nothing happens."
            abc.xyzzy()
        `, 16)
		abc, _ := vm.Get("abc")
		def, _ := vm.Get("def")
		object := abc.Object()
		result, _ := object.Call("xyzzy")
		is(result, 16)
		result, _ = object.Call("xyzzy", 1)
		is(result, 17)
		value, _ := object.Get("xyzzy")
		result, _ = value.Call(def)
		is(result, 27)
		result, _ = value.Call(def, 3)
		is(result, 30)
		object = value.Object() // Object xyzzy
		result, _ = object.Value().Call(def, 3)
		is(result, 30)

		test(`
            abc = {
                'abc': 1,
                'def': false,
                3.14159: NaN,
            };
            abc['abc'];
        `, 1)
		abc, err := vm.Get("abc")
		is(err, nil)
		object = abc.Object() // Object abc
		value, err = object.Get("abc")
		is(err, nil)
		is(value, 1)
		is(object.Keys(), []string{"abc", "def", "3.14159"})

		test(`
            abc = [ 0, 1, 2, 3.14159, "abc", , ];
            abc.def = true;
        `)
		abc, err = vm.Get("abc")
		is(err, nil)
		object = abc.Object() // Object abc
		is(object.Keys(), []string{"0", "1", "2", "3", "4", "def"})
	})
}

func TestObjectKeys(t *testing.T) {
	tt(t, func() {
		vm := New()
		vm.Eval(`var x = Object.create(null); x.a = 1`)
		vm.Eval(`var y = Object.create(x); y.b = 2`)

		o1, _ := vm.Object("x")
		is(o1.Keys(), []string{"a"})
		is(o1.KeysByParent(), [][]string{{"a"}})

		o2, _ := vm.Object("y")
		is(o2.Keys(), []string{"b"})
		is(o2.KeysByParent(), [][]string{{"b"}, {"a"}})
	})
}

func TestUnicode(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`var abc = eval("\"a\uFFFFa\"");`, "undefined")

		test(`abc.length`, 3)

		test(`abc != "aa"`, true)

		test("abc[1] === \"\uFFFF\"", true)
	})
}

func TestDotMember(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`
            abc = {
                ghi: 11,
            }
            abc.def = "Xyzzy"
            abc.null = "Nothing happens."
        `)
		test(`abc.def`, "Xyzzy")
		test(`abc.null`, "Nothing happens.")
		test(`abc.ghi`, 11)

		test(`
            abc = {
                null: 11,
            }
        `)
		test(`abc.def`, "undefined")
		test(`abc.null`, 11)
		test(`abc.ghi`, "undefined")
	})
}

func Test_stringToFloat(t *testing.T) {
	tt(t, func() {

		is(parseNumber("10e10000"), _Infinity)
		is(parseNumber("10e10_."), _NaN)
	})
}

func Test_delete(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`
            delete 42;
        `, true)

		test(`
            var abc = delete $_undefined_$;
            abc = abc && delete ($_undefined_$);
            abc;
        `, true)

		// delete should not trigger get()
		test(`
            var abc = {
                get def() {
                    throw "Test_delete: delete should not trigger get()"
                }
            };
            delete abc.def
        `, true)
	})
}

func TestObject_defineOwnProperty(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`
            var object = {};

            var descriptor = new Boolean(false);
            descriptor.configurable = true;

            Object.defineProperties(object, {
                property: descriptor
            });

            var abc = object.hasOwnProperty("property");
            delete object.property;
            var def = object.hasOwnProperty("property");

            [ abc, def ];
        `, "true,false")

		test(`
            var object = [0, 1, 2];
            Object.defineProperty(object, "0", {
                value: 42,
                writable: false,
                enumerable: false,
                configurable: false
            });
            var abc = Object.getOwnPropertyDescriptor(object, "0");
            [ abc.value, abc.writable, abc.enumerable, abc.configurable ];
        `, "42,false,false,false")

		test(`
            var abc = { "xyzzy": 42 };
            var def = Object.defineProperties(abc, "");
            abc === def;
        `, true)
	})
}

func Test_assignmentEvaluationOrder(t *testing.T) {
	tt(t, func() {
		test, _ := test()

		test(`
            var abc = 0;
            ((abc = 1) & abc);
        `, 1)

		test(`
            var abc = 0;
            (abc & (abc = 1));
        `, 0)
	})
}

func TestOttoCall(t *testing.T) {
	tt(t, func() {
		vm := New()

		_, err := vm.Run(`
            var abc = {
                ghi: 1,
                def: function(def){
                    var ghi = 0;
                    if (this.ghi) {
                        ghi = this.ghi;
                    }
                    return "def: " + (def + 3.14159 + ghi);
                }
            };
            function structFunc(s) {
                return s.Val;
            }
        `)
		is(err, nil)

		value, err := vm.Call(`abc.def`, nil, 2)
		is(err, nil)
		is(value, "def: 6.14159")

		value, err = vm.Call(`abc.def`, "", 2)
		is(err, nil)
		is(value, "def: 5.14159")

		// Do not attempt to do a ToValue on a this of nil
		value, err = vm.Call(`jkl.def`, nil, 1, 2, 3)
		is(err, "!=", nil)
		is(value, "undefined")

		value, err = vm.Call(`[ 1, 2, 3, undefined, 4 ].concat`, nil, 5, 6, 7, "abc")
		is(err, nil)
		is(value, "1,2,3,,4,5,6,7,abc")

		s := struct{ Val int }{Val: 10}
		value, err = vm.Call("structFunc", nil, s)
		is(err, nil)
		is(value, 10)
	})
}

func TestOttoCall_new(t *testing.T) {
	tt(t, func() {
		test, vm := test()

		vm.Set("abc", func(call FunctionCall) Value {
			value, err := call.Otto.Call(`new Object`, nil, "Nothing happens.")
			is(err, nil)
			return value
		})
		test(`
            def = abc();
            [ def, def instanceof String ];
        `, "Nothing happens.,true")
	})
}

func TestOttoCall_newWithBrackets(t *testing.T) {
	tt(t, func() {
		test, vm := test()

		_, err := vm.Run(`var a = {default: function B(x) { this.x = x; } }`)
		is(err, nil)

		test(`(new a['default'](1)).x`, 1)
	})
}

func TestOttoCall_throw(t *testing.T) {
	// FIXME? (Been broken for a while)
	// Looks like this has been broken for a while... what
	// behavior do we want here?

	if true {
		return
	}

	tt(t, func() {
		test, vm := test()

		vm.Set("abc", func(call FunctionCall) Value {
			if false {
				call.Otto.Call(`throw eval`, nil, "({ def: 3.14159 })")
			}
			call.Otto.Call(`throw Error`, nil, "abcdef")
			return Value{}
		})
		// TODO try { abc(); } catch (err) { error = err }
		// Possible unrelated error case:
		// If error is not declared beforehand, is later referencing it a ReferenceError?
		// Should the catch { } declare error in the outer scope?
		test(`
            var error;
            try {
                abc();
            }
            catch (err) {
                error = err;
            }
            [ error instanceof Error, error.message, error.def ];
        `, "true,abcdef,")

		vm.Set("def", func(call FunctionCall) Value {
			call.Otto.Call(`throw new Object`, nil, 3.14159)
			return UndefinedValue()
		})
		test(`
            try {
                def();
            }
            catch (err) {
                error = err;
            }
            [ error instanceof Error, error.message, error.def, typeof error, error, error instanceof Number ];
        `, "false,,,object,3.14159,true")
	})
}

func TestOttoCopy(t *testing.T) {
	tt(t, func() {
		vm0 := New()
		vm0.Run(`
            var abc = function() {
                return "Xyzzy";
            };

            function def() {
                return abc() + (0 + {});
            }
        `)

		value, err := vm0.Run(`
            def();
        `)
		is(err, nil)
		is(value, "Xyzzy0[object Object]")

		vm1 := vm0.Copy()
		value, err = vm1.Run(`
            def();
        `)
		is(err, nil)
		is(value, "Xyzzy0[object Object]")

		vm1.Run(`
            abc = function() {
                return 3.14159;
            };
        `)
		value, err = vm1.Run(`
            def();
        `)
		is(err, nil)
		is(value, "3.141590[object Object]")

		value, err = vm0.Run(`
            def();
        `)
		is(err, nil)
		is(value, "Xyzzy0[object Object]")

		{
			vm0 := New()
			vm0.Run(`
                var global = (function () {return this;}())
                var abc = 0;
                var vm = "vm0";

                var def = (function(){
                    var jkl = 0;
                    var abc = function() {
                        global.abc += 1;
                        jkl += 1;
                        return 1;
                    };

                    return function() {
                        return [ vm, global.abc, jkl, abc() ];
                    };
                })();
            `)

			value, err := vm0.Run(`
                def();
            `)
			is(err, nil)
			is(value, "vm0,0,0,1")

			vm1 := vm0.Copy()
			vm1.Set("vm", "vm1")
			value, err = vm1.Run(`
                def();
            `)
			is(err, nil)
			is(value, "vm1,1,1,1")

			value, err = vm0.Run(`
                def();
            `)
			is(err, nil)
			is(value, "vm0,1,1,1")

			value, err = vm1.Run(`
                def();
            `)
			is(err, nil)
			is(value, "vm1,2,2,1")
		}
	})
}

func TestOttoCall_clone(t *testing.T) {
	tt(t, func() {
		vm := New().clone()
		rt := vm.runtime

		{
			// FIXME terst, Check how this comparison is done
			is(rt.global.Array.prototype, rt.global.FunctionPrototype)
			is(rt.global.ArrayPrototype, "!=", nil)
			is(rt.global.Array.runtime, rt)
			is(rt.global.Array.prototype.runtime, rt)
			is(rt.global.Array.get("prototype")._object().runtime, rt)
		}

		{
			value, err := vm.Run(`[ 1, 2, 3 ].toString()`)
			is(err, nil)
			is(value, "1,2,3")
		}

		{
			value, err := vm.Run(`[ 1, 2, 3 ]`)
			is(err, nil)
			is(value, "1,2,3")
			object := value._object()
			is(object, "!=", nil)
			is(object.prototype, rt.global.ArrayPrototype)

			value, err = vm.Run(`Array.prototype`)
			is(err, nil)
			object = value._object()
			is(object.runtime, rt)
			is(object, "!=", nil)
			is(object, rt.global.ArrayPrototype)
		}

		{
			otto1 := New()
			_, err := otto1.Run(`
                var abc = 1;
                var def = 2;
            `)
			is(err, nil)

			otto2 := otto1.clone()
			value, err := otto2.Run(`abc += 1; abc;`)
			is(err, nil)
			is(value, 2)

			value, err = otto1.Run(`abc += 4; abc;`)
			is(err, nil)
			is(value, 5)
		}

		{
			vm1 := New()
			_, err := vm1.Run(`
                var abc = 1;
                var def = function(value) {
                    abc += value;
                    return abc;
                }
            `)
			is(err, nil)

			vm2 := vm1.clone()
			value, err := vm2.Run(`def(1)`)
			is(err, nil)
			is(value, 2)

			value, err = vm1.Run(`def(4)`)
			is(err, nil)
			is(value, 5)
		}

		{
			vm1 := New()
			_, err := vm1.Run(`
                var abc = {
                    ghi: 1,
                    jkl: function(value) {
                        this.ghi += value;
                        return this.ghi;
                    }
                };
                var def = {
                    abc: abc
                };
            `)
			is(err, nil)

			otto2 := vm1.clone()
			value, err := otto2.Run(`def.abc.jkl(1)`)
			is(err, nil)
			is(value, 2)

			value, err = vm1.Run(`def.abc.jkl(4)`)
			is(err, nil)
			is(value, 5)
		}

		{
			vm1 := New()
			_, err := vm1.Run(`
                var abc = function() { return "abc"; };
                var def = function() { return "def"; };
            `)
			is(err, nil)

			vm2 := vm1.clone()
			value, err := vm2.Run(`
                [ abc.toString(), def.toString() ];
            `)
			is(value, `function() { return "abc"; },function() { return "def"; }`)

			_, err = vm2.Run(`
                var def = function() { return "ghi"; };
            `)
			is(err, nil)

			value, err = vm1.Run(`
                [ abc.toString(), def.toString() ];
            `)
			is(value, `function() { return "abc"; },function() { return "def"; }`)

			value, err = vm2.Run(`
                [ abc.toString(), def.toString() ];
            `)
			is(value, `function() { return "abc"; },function() { return "ghi"; }`)
		}

	})
}

func TestOttoRun(t *testing.T) {
	tt(t, func() {
		vm := New()

		program, err := parser.ParseFile(nil, "", "", 0)
		is(err, nil)
		value, err := vm.Run(program)
		is(err, nil)
		is(value, UndefinedValue())

		program, err = parser.ParseFile(nil, "", "2 + 2", 0)
		is(err, nil)
		value, err = vm.Run(program)
		is(err, nil)
		is(value, 4)
		value, err = vm.Run(program)
		is(err, nil)
		is(value, 4)

		program, err = parser.ParseFile(nil, "", "var abc; if (!abc) abc = 0; abc += 2; abc;", 0)
		value, err = vm.Run(program)
		is(err, nil)
		is(value, 2)
		value, err = vm.Run(program)
		is(err, nil)
		is(value, 4)
		value, err = vm.Run(program)
		is(err, nil)
		is(value, 6)

		{
			src := []byte("var abc; if (!abc) abc = 0; abc += 2; abc;")
			value, err = vm.Run(src)
			is(err, nil)
			is(value, 8)

			value, err = vm.Run(bytes.NewBuffer(src))
			is(err, nil)
			is(value, 10)

			value, err = vm.Run(io.Reader(bytes.NewBuffer(src)))
			is(err, nil)
			is(value, 12)
		}

		{
			script, err := vm.Compile("", `var abc; if (!abc) abc = 0; abc += 2; abc;`)
			is(err, nil)

			value, err = vm.Run(script)
			is(err, nil)
			is(value, 14)

			value, err = vm.Run(script)
			is(err, nil)
			is(value, 16)

			is(script.String(), "// \nvar abc; if (!abc) abc = 0; abc += 2; abc;")
		}
	})
}

// This generates functions to be used by the test below. The arguments are
// `src`, which is something that otto can execute, and `expected`, which is
// what the result of executing `src` should be.
func makeTestOttoEvalFunction(src, expected interface{}) func(c FunctionCall) Value {
	return func(c FunctionCall) Value {
		v, err := c.Otto.Eval(src)
		is(err, nil)
		if err != nil {
			panic(err)
		}

		i, err := v.Export()
		is(err, nil)
		if err != nil {
			panic(err)
		}

		is(i, expected)

		return v
	}
}

func TestOttoEval(t *testing.T) {
	tt(t, func() {
		vm := New()

		vm.Set("x1", makeTestOttoEvalFunction(`a`, 1))
		vm.Set("y1", makeTestOttoEvalFunction(`b`, "hello"))
		vm.Set("z1", makeTestOttoEvalFunction(`c`, true))
		vm.Set("w", makeTestOttoEvalFunction(`a = 2; b = 'what'; c = false; null`, nil))
		vm.Set("x2", makeTestOttoEvalFunction(`a`, 2))
		vm.Set("y2", makeTestOttoEvalFunction(`b`, "what"))
		vm.Set("z2", makeTestOttoEvalFunction(`c`, false))

		// note that these variables are defined in the scope of function `t`,
		// so would not usually be available to the functions called below.
		//
		// this is _not_ the recommended use case for `Eval` - instead it's
		// intended to be used in `debugger` handlers. this code here is the
		// equivalent of reading behind the current stack frame in C...
		// technically valid, but completely insane.
		//
		// makes for a good test case though.
		_, err := vm.Run(`(function t() {
            var a = 1;
            var b = 'hello';
            var c = true;

            x1();
            y1();
            z1();
            w();
            x2();
            y2();
            z2();
        }())`)

		is(err, nil)
	})

	// this test makes sure that `Eval` doesn't explode if the VM doesn't have
	// a scope other than global defined.
	tt(t, func() {
		vm := New()

		_, err := vm.Eval("null")
		is(err, nil)

		vm.Set("a", 1)
		vm.Set("b", 2)

		v, err := vm.Eval("a + b")
		is(err, nil)
		r, err := v.Export()
		is(err, nil)
		is(r, 3)
	})
}

func TestOttoContext(t *testing.T) {
	// These are all the builtin global scope symbols
	builtins := []string{
		"escape",
		"URIError",
		"RegExp",
		"ReferenceError",
		"parseFloat",
		"parseInt",
		"SyntaxError",
		"decodeURIComponent",
		"encodeURIComponent",
		"Infinity",
		"JSON",
		"isNaN",
		"unescape",
		"decodeURI",
		"Object",
		"Function",
		"RangeError",
		"Error",
		"get_context",
		"eval",
		"Number",
		"Math",
		"NaN",
		"Date",
		"Boolean",
		"console",
		"encodeURI",
		"EvalError",
		"Array",
		"TypeError",
		"String",
		"isFinite",
		"undefined",
	}

	tt(t, func() {
		vm := New()

		vm.Set("get_context", func(c FunctionCall) Value {
			ctx := c.Otto.Context()
			is(ctx.Callee, "f1")
			is(ctx.Filename, "<anonymous>")
			is(ctx.Line, 8)
			is(ctx.Column, 5)
			is(ctx.Stacktrace, []string{
				"f1 (<anonymous>:8:5)",
				"f2 (<anonymous>:15:5)",
				"f3 (<anonymous>:19:5)",
				"t (<anonymous>:22:4)",
			})
			is(len(ctx.Symbols), 9+len(builtins))
			is(ctx.Symbols["a"], 1)
			is(ctx.Symbols["b"], "hello")
			is(ctx.Symbols["c"], true)
			is(ctx.Symbols["j"], 2)
			is(ctx.Symbols["f1"].IsFunction(), true)
			is(ctx.Symbols["f2"].IsFunction(), true)
			is(ctx.Symbols["f3"].IsFunction(), true)
			is(ctx.Symbols["t"].IsFunction(), true)
			callee, _ := ctx.Symbols["arguments"].Object().Get("callee")
			is(callee.IsDefined(), true)

			return Value{}
		})

		_, err := vm.Run(`(function t() {
			var a = 1;
			var b = 'hello';
			var c = true;

			function f1() {
				var j = 2;
				get_context();
				(function() {
					var d = 4;
				})()
			}

			function f2() {
				f1();
			}

			function f3() {
				f2();
			}

			f3();

			a = 2;
			b = 'goodbye';
			c = false;
		}())`)

		is(err, nil)
	})

	// this test makes sure that `Context` works on global scope by default, if
	// there is not a current scope.
	tt(t, func() {
		vm := New()

		vm.Set("get_context", func(c FunctionCall) Value {
			ctx := c.Otto.Context()
			is(ctx.Callee, "")
			is(ctx.Filename, "<anonymous>")
			is(ctx.Line, 3)
			is(ctx.Column, 4)
			is(ctx.Stacktrace, []string{"<anonymous>:3:4"})
			is(len(ctx.Symbols), 2+len(builtins))
			is(ctx.Symbols["a"], 1)
			is(ctx.Symbols["b"], UndefinedValue())

			return Value{}
		})

		_, err := vm.Run(`
			var a = 1;
			get_context()
			var b = 2;
		`)
		is(err, nil)
	})

	// this test makes sure variables are shadowed correctly.
	tt(t, func() {
		vm := New()

		vm.Set("check_context", func(c FunctionCall) Value {
			n, err := c.Argument(0).ToInteger()
			is(err, nil)

			ctx := c.Otto.Context()
			is(ctx.Symbols["a"], n)

			return Value{}
		})

		_, err := vm.Run(`
            var a = 1;
            check_context(1);
            (function() {
                var a = 2;
                check_context(2);
            }());
            (function(a) {
                check_context(3);
            }(3));
            (function(a) {
                check_context(4);
            }).call(null, 4);
            check_context(1);
        `)
		is(err, nil)
	})
}

func Test_objectLength(t *testing.T) {
	tt(t, func() {
		_, vm := test()

		value := vm.Set("abc", []string{"jkl", "mno"})
		is(objectLength(value._object()), 2)

		value, _ = vm.Run(`[1, 2, 3]`)
		is(objectLength(value._object()), 3)

		value, _ = vm.Run(`new String("abcdefghi")`)
		is(objectLength(value._object()), 9)

		value, _ = vm.Run(`"abcdefghi"`)
		is(objectLength(value._object()), 0)
	})
}

func Test_stackLimit(t *testing.T) {
	// JavaScript stack depth before entering `a` is 5; becomes 6 after
	// entering. setting the maximum stack depth to 5 should result in an
	// error ocurring at that 5 -> 6 boundary.
	code := `
        function a() {}
        function b() { a(); }
        function c() { b(); }
        function d() { c(); }
        function e() { d(); }
        e();
    `

	// has no error
	tt(t, func() {
		_, vm := test()

		_, err := vm.Run(code)

		is(err == nil, true)
	})

	// has error
	tt(t, func() {
		_, vm := test()

		vm.vm.SetStackDepthLimit(2)

		_, err := vm.Run(code)

		is(err == nil, false)
	})

	// has error
	tt(t, func() {
		_, vm := test()

		vm.vm.SetStackDepthLimit(5)

		_, err := vm.Run(code)

		is(err == nil, false)
	})

	// has no error
	tt(t, func() {
		_, vm := test()

		vm.vm.SetStackDepthLimit(6)

		_, err := vm.Run(code)

		is(err == nil, true)
	})

	// has no error
	tt(t, func() {
		_, vm := test()

		vm.vm.SetStackDepthLimit(1)
		vm.vm.SetStackDepthLimit(0)

		_, err := vm.Run(code)

		is(err == nil, true)
	})
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New()
	}
}

func BenchmarkClone(b *testing.B) {
	vm := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm.clone()
	}
}
