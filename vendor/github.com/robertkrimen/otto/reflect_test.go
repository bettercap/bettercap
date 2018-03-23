package otto

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type _abcStruct struct {
	Abc bool
	Def int
	Ghi string
	Jkl interface{}
	Mno _mnoStruct
	Pqr map[string]int8
}

func (abc _abcStruct) String() string {
	return abc.Ghi
}

func (abc *_abcStruct) FuncPointer() string {
	return "abc"
}

func (abc _abcStruct) Func() {
	return
}

func (abc _abcStruct) FuncReturn1() string {
	return "abc"
}

func (abc _abcStruct) FuncReturn2() (string, error) {
	return "def", nil
}

func (abc _abcStruct) Func1Return1(a string) string {
	return a
}

func (abc _abcStruct) Func2Return1(x, y string) string {
	return x + y
}

func (abc _abcStruct) FuncEllipsis(xyz ...string) int {
	return len(xyz)
}

func (abc _abcStruct) FuncReturnStruct() _mnoStruct {
	return _mnoStruct{}
}

func (abs _abcStruct) Func1Int(i int) int {
	return i + 1
}

func (abs _abcStruct) Func1Int8(i int8) int8 {
	return i + 1
}

func (abs _abcStruct) Func1Int16(i int16) int16 {
	return i + 1
}

func (abs _abcStruct) Func1Int32(i int32) int32 {
	return i + 1
}

func (abs _abcStruct) Func1Int64(i int64) int64 {
	return i + 1
}

func (abs _abcStruct) Func1Uint(i uint) uint {
	return i + 1
}

func (abs _abcStruct) Func1Uint8(i uint8) uint8 {
	return i + 1
}

func (abs _abcStruct) Func1Uint16(i uint16) uint16 {
	return i + 1
}

func (abs _abcStruct) Func1Uint32(i uint32) uint32 {
	return i + 1
}

func (abs _abcStruct) Func1Uint64(i uint64) uint64 {
	return i + 1
}

func (abs _abcStruct) Func2Int(i, j int) int {
	return i + j
}

func (abs _abcStruct) Func2StringInt(s string, i int) string {
	return fmt.Sprintf("%v:%v", s, i)
}

func (abs _abcStruct) Func1IntVariadic(a ...int) int {
	t := 0
	for _, i := range a {
		t += i
	}
	return t
}

func (abs _abcStruct) Func2IntVariadic(s string, a ...int) string {
	t := 0
	for _, i := range a {
		t += i
	}
	return fmt.Sprintf("%v:%v", s, t)
}

func (abs _abcStruct) Func2IntArrayVariadic(s string, a ...[]int) string {
	t := 0
	for _, i := range a {
		for _, j := range i {
			t += j
		}
	}
	return fmt.Sprintf("%v:%v", s, t)
}

type _mnoStruct struct {
	Ghi string
}

func (mno _mnoStruct) Func() string {
	return "mno"
}

func TestReflect(t *testing.T) {
	if true {
		return
	}
	tt(t, func() {
		// Testing dbgf
		// These should panic
		toValue("Xyzzy").toReflectValue(reflect.Ptr)
		stringToReflectValue("Xyzzy", reflect.Ptr)
	})
}

func Test_reflectStruct(t *testing.T) {
	tt(t, func() {
		test, vm := test()

		// _abcStruct
		{
			abc := &_abcStruct{}
			vm.Set("abc", abc)

			test(`
                [ abc.Abc, abc.Ghi ];
            `, "false,")

			abc.Abc = true
			abc.Ghi = "Nothing happens."

			test(`
                [ abc.Abc, abc.Ghi ];
            `, "true,Nothing happens.")

			*abc = _abcStruct{}

			test(`
                [ abc.Abc, abc.Ghi ];
            `, "false,")

			abc.Abc = true
			abc.Ghi = "Xyzzy"
			vm.Set("abc", abc)

			test(`
                [ abc.Abc, abc.Ghi ];
            `, "true,Xyzzy")

			is(abc.Abc, true)
			test(`
                abc.Abc = false;
                abc.Def = 451;
                abc.Ghi = "Nothing happens.";
                abc.abc = "Something happens.";
                [ abc.Def, abc.abc ];
            `, "451,Something happens.")
			is(abc.Abc, false)
			is(abc.Def, 451)
			is(abc.Ghi, "Nothing happens.")

			test(`
                delete abc.Def;
                delete abc.abc;
                [ abc.Def, abc.abc ];
            `, "451,")
			is(abc.Def, 451)

			test(`
                abc.FuncPointer();
            `, "abc")

			test(`
                abc.Func();
            `, "undefined")

			test(`
                abc.FuncReturn1();
            `, "abc")

			test(`
                abc.Func1Return1("abc");
            `, "abc")

			test(`
                abc.Func2Return1("abc", "def");
            `, "abcdef")

			test(`
                abc.FuncEllipsis("abc", "def", "ghi");
            `, 3)

			test(`
                ret = abc.FuncReturn2();
                if (ret && ret.length && ret.length == 2 && ret[0] == "def" && ret[1] === undefined) {
                        true;
                } else {
                       false;
                }
            `, true)

			test(`
                abc.FuncReturnStruct();
            `, "[object Object]")

			test(`
                abc.FuncReturnStruct().Func();
            `, "mno")

			test(`
                abc.Func1Int(1);
            `, 2)

			test(`
                abc.Func1Int(0x01 & 0x01);
            `, 2)

			test(`raise:
                abc.Func1Int(1.1);
            `, "RangeError: converting float64 to int would cause loss of precision")

			test(`
		var v = 1;
                abc.Func1Int(v + 1);
            `, 3)

			test(`
                abc.Func2Int(1, 2);
            `, 3)

			test(`
                abc.Func1Int8(1);
            `, 2)

			test(`
                abc.Func1Int16(1);
            `, 2)

			test(`
                abc.Func1Int32(1);
            `, 2)

			test(`
                abc.Func1Int64(1);
            `, 2)

			test(`
                abc.Func1Uint(1);
            `, 2)

			test(`
                abc.Func1Uint8(1);
            `, 2)

			test(`
                abc.Func1Uint16(1);
            `, 2)

			test(`
                abc.Func1Uint32(1);
            `, 2)

			test(`
                abc.Func1Uint64(1);
            `, 2)

			test(`
                abc.Func2StringInt("test", 1);
            `, "test:1")

			test(`
                abc.Func1IntVariadic(1, 2);
            `, 3)

			test(`
                abc.Func2IntVariadic("test", 1, 2);
            `, "test:3")

			test(`
                abc.Func2IntVariadic("test", [1, 2]);
            `, "test:3")

			test(`
                abc.Func2IntArrayVariadic("test", [1, 2]);
            `, "test:3")

			test(`
                abc.Func2IntArrayVariadic("test", [1, 2], [3, 4]);
            `, "test:10")

			test(`
                abc.Func2IntArrayVariadic("test", [[1, 2], [3, 4]]);
            `, "test:10")
		}
	})
}

func Test_reflectMap(t *testing.T) {
	tt(t, func() {
		test, vm := test()

		// map[string]string
		{
			abc := map[string]string{
				"Xyzzy": "Nothing happens.",
				"def":   "1",
			}
			vm.Set("abc", abc)

			test(`
                abc.xyz = "pqr";
                [ abc.Xyzzy, abc.def, abc.ghi ];
            `, "Nothing happens.,1,")

			is(abc["xyz"], "pqr")
		}

		// map[string]float64
		{
			abc := map[string]float64{
				"Xyzzy": math.Pi,
				"def":   1,
			}
			vm.Set("abc", abc)

			test(`
                abc.xyz = "pqr";
                abc.jkl = 10;
                [ abc.Xyzzy, abc.def, abc.ghi ];
            `, "3.141592653589793,1,")

			is(abc["xyz"], math.NaN())
			is(abc["jkl"], float64(10))
		}

		// map[string]int32
		{
			abc := map[string]int32{
				"Xyzzy": 3,
				"def":   1,
			}
			vm.Set("abc", abc)

			test(`
                abc.xyz = "pqr";
                abc.jkl = 10;
                [ abc.Xyzzy, abc.def, abc.ghi ];
            `, "3,1,")

			is(abc["xyz"], 0)
			is(abc["jkl"], int32(10))

			test(`
                delete abc["Xyzzy"];
            `)

			_, exists := abc["Xyzzy"]
			is(exists, false)
			is(abc["Xyzzy"], 0)
		}

		// map[int32]string
		{
			abc := map[int32]string{
				0: "abc",
				1: "def",
			}
			vm.Set("abc", abc)

			test(`
                abc[2] = "pqr";
                //abc.jkl = 10;
                abc[3] = 10;
                [ abc[0], abc[1], abc[2], abc[3] ]
            `, "abc,def,pqr,10")

			is(abc[2], "pqr")
			is(abc[3], "10")

			test(`
                delete abc[2];
            `)

			_, exists := abc[2]
			is(exists, false)
		}

	})
}

func Test_reflectMapIterateKeys(t *testing.T) {
	tt(t, func() {
		test, vm := test()

		// map[string]interface{}
		{
			abc := map[string]interface{}{
				"Xyzzy": "Nothing happens.",
				"def":   1,
			}
			vm.Set("abc", abc)
			test(`
                var keys = [];
                for (var key in abc) {
                  keys.push(key);
                }
                keys.sort();
                keys;
            `, "Xyzzy,def")
		}

		// map[uint]interface{}
		{
			abc := map[uint]interface{}{
				456: "Nothing happens.",
				123: 1,
			}
			vm.Set("abc", abc)
			test(`
                var keys = [];
                for (var key in abc) {
                  keys.push(key);
                }
                keys.sort();
                keys;
            `, "123,456")
		}

		// map[byte]interface{}
		{
			abc := map[byte]interface{}{
				10: "Nothing happens.",
				20: 1,
			}
			vm.Set("abc", abc)
			test(`
                for (var key in abc) {
                  abc[key] = "123";
                }
            `)
			is(abc[10], "123")
			is(abc[20], "123")
		}

	})
}

func Test_reflectSlice(t *testing.T) {
	tt(t, func() {
		test, vm := test()

		// []bool
		{
			abc := []bool{
				false,
				true,
				true,
				false,
			}
			vm.Set("abc", abc)

			test(`
                abc;
            `, "false,true,true,false")

			test(`
                abc[0] = true;
                abc[abc.length-1] = true;
                delete abc[2];
                abc;
            `, "true,true,false,true")

			is(abc, []bool{true, true, false, true})
			is(abc[len(abc)-1], true)
		}

		// []int32
		{
			abc := make([]int32, 4)
			vm.Set("abc", abc)

			test(`
                abc;
            `, "0,0,0,0")

			test(`raise:
                abc[0] = "42";
                abc[1] = 4.2;
                abc[2] = 3.14;
                abc;
            `, "RangeError: 4.2 to reflect.Kind: int32")

			is(abc, []int32{42, 0, 0, 0})

			test(`
                delete abc[1];
                delete abc[2];
            `)
			is(abc[1], 0)
			is(abc[2], 0)
		}
	})
}

func Test_reflectArray(t *testing.T) {
	tt(t, func() {
		test, vm := test()

		// []bool
		{
			abc := [4]bool{
				false,
				true,
				true,
				false,
			}
			vm.Set("abc", abc)

			test(`
                abc;
            `, "false,true,true,false")
			// Unaddressable array

			test(`
                abc[0] = true;
                abc[abc.length-1] = true;
                abc;
            `, "false,true,true,false")
			// Again, unaddressable array

			is(abc, [4]bool{false, true, true, false})
			is(abc[len(abc)-1], false)
			// ...
		}
		// []int32
		{
			abc := make([]int32, 4)
			vm.Set("abc", abc)

			test(`
                abc;
            `, "0,0,0,0")

			test(`raise:
                abc[0] = "42";
                abc[1] = 4.2;
                abc[2] = 3.14;
                abc;
            `, "RangeError: 4.2 to reflect.Kind: int32")

			is(abc, []int32{42, 0, 0, 0})
		}

		// []bool
		{
			abc := [4]bool{
				false,
				true,
				true,
				false,
			}
			vm.Set("abc", &abc)

			test(`
                abc;
            `, "false,true,true,false")

			test(`
                abc[0] = true;
                abc[abc.length-1] = true;
                delete abc[2];
                abc;
            `, "true,true,false,true")

			is(abc, [4]bool{true, true, false, true})
			is(abc[len(abc)-1], true)
		}

		// no common type
		{
			test(`
                 abc = [1, 2.2, "str"];
                 abc;
             `, "1,2.2,str")
			val, err := vm.Get("abc")
			is(err, nil)
			abc, err := val.Export()
			is(err, nil)
			is(abc, []interface{}{int64(1), 2.2, "str"})
		}

		// common type int
		{
			test(`
                 abc = [1, 2, 3];
                 abc;
             `, "1,2,3")
			val, err := vm.Get("abc")
			is(err, nil)
			abc, err := val.Export()
			is(err, nil)
			is(abc, []int64{1, 2, 3})
		}

		// common type string
		{

			test(`
                 abc = ["str1", "str2", "str3"];
                 abc;
             `, "str1,str2,str3")

			val, err := vm.Get("abc")
			is(err, nil)
			abc, err := val.Export()
			is(err, nil)
			is(abc, []string{"str1", "str2", "str3"})
		}

		// issue #269
		{
			called := false
			vm.Set("blah", func(c FunctionCall) Value {
				v, err := c.Argument(0).Export()
				is(err, nil)
				is(v, []int64{3})
				called = true
				return UndefinedValue()
			})
			is(called, false)
			test(`var x = 3; blah([x])`)
			is(called, true)
		}
	})
}

func Test_reflectArray_concat(t *testing.T) {
	tt(t, func() {
		test, vm := test()

		vm.Set("ghi", []string{"jkl", "mno"})
		vm.Set("pqr", []interface{}{"jkl", 42, 3.14159, true})
		test(`
            var def = {
                "abc": ["abc"],
                "xyz": ["xyz"]
            };
            xyz = pqr.concat(ghi, def.abc, def, def.xyz);
            [ xyz, xyz.length ];
        `, "jkl,42,3.14159,true,jkl,mno,abc,[object Object],xyz,9")
	})
}

func Test_reflectMapInterface(t *testing.T) {
	tt(t, func() {
		test, vm := test()

		{
			abc := map[string]interface{}{
				"Xyzzy": "Nothing happens.",
				"def":   "1",
				"jkl":   "jkl",
			}
			vm.Set("abc", abc)
			vm.Set("mno", &_abcStruct{})

			test(`
                abc.xyz = "pqr";
                abc.ghi = {};
                abc.jkl = 3.14159;
                abc.mno = mno;
                mno.Abc = true;
                mno.Ghi = "Something happens.";
                [ abc.Xyzzy, abc.def, abc.ghi, abc.mno ];
            `, "Nothing happens.,1,[object Object],[object Object]")

			is(abc["xyz"], "pqr")
			is(abc["ghi"], "[object Object]")
			is(abc["jkl"], float64(3.14159))
			mno, valid := abc["mno"].(*_abcStruct)
			is(valid, true)
			is(mno.Abc, true)
			is(mno.Ghi, "Something happens.")
		}
	})
}

func TestPassthrough(t *testing.T) {
	tt(t, func() {
		test, vm := test()

		{
			abc := &_abcStruct{
				Mno: _mnoStruct{
					Ghi: "<Mno.Ghi>",
				},
			}
			vm.Set("abc", abc)

			test(`
                abc.Mno.Ghi;
            `, "<Mno.Ghi>")

			vm.Set("pqr", map[string]int8{
				"xyzzy":            0,
				"Nothing happens.": 1,
			})

			test(`
                abc.Ghi = "abc";
                abc.Pqr = pqr;
                abc.Pqr["Nothing happens."];
            `, 1)

			mno := _mnoStruct{
				Ghi: "<mno.Ghi>",
			}
			vm.Set("mno", mno)

			test(`
                abc.Mno = mno;
                abc.Mno.Ghi;
            `, "<mno.Ghi>")
		}
	})
}
