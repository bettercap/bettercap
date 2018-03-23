package otto

import (
	"reflect"
	"testing"
)

func BenchmarkNativeCallWithString(b *testing.B) {
	vm := New()
	vm.Set("x", func(a1 string) {})
	s, _ := vm.Compile("test.js", `x("zzz")`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithFloat32(b *testing.B) {
	vm := New()
	vm.Set("x", func(a1 float32) {})
	s, _ := vm.Compile("test.js", `x(1.1)`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithFloat64(b *testing.B) {
	vm := New()
	vm.Set("x", func(a1 float64) {})
	s, _ := vm.Compile("test.js", `x(1.1)`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithInt(b *testing.B) {
	vm := New()
	vm.Set("x", func(a1 int) {})
	s, _ := vm.Compile("test.js", `x(1)`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithUint(b *testing.B) {
	vm := New()
	vm.Set("x", func(a1 uint) {})
	s, _ := vm.Compile("test.js", `x(1)`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithInt8(b *testing.B) {
	vm := New()
	vm.Set("x", func(a1 int8) {})
	s, _ := vm.Compile("test.js", `x(1)`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithUint8(b *testing.B) {
	vm := New()
	vm.Set("x", func(a1 uint8) {})
	s, _ := vm.Compile("test.js", `x(1)`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithInt16(b *testing.B) {
	vm := New()
	vm.Set("x", func(a1 int16) {})
	s, _ := vm.Compile("test.js", `x(1)`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithUint16(b *testing.B) {
	vm := New()
	vm.Set("x", func(a1 uint16) {})
	s, _ := vm.Compile("test.js", `x(1)`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithInt32(b *testing.B) {
	vm := New()
	vm.Set("x", func(a1 int32) {})
	s, _ := vm.Compile("test.js", `x(1)`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithUint32(b *testing.B) {
	vm := New()
	vm.Set("x", func(a1 uint32) {})
	s, _ := vm.Compile("test.js", `x(1)`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithInt64(b *testing.B) {
	vm := New()
	vm.Set("x", func(a1 int64) {})
	s, _ := vm.Compile("test.js", `x(1)`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithUint64(b *testing.B) {
	vm := New()
	vm.Set("x", func(a1 uint64) {})
	s, _ := vm.Compile("test.js", `x(1)`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithStringInt(b *testing.B) {
	vm := New()
	vm.Set("x", func(a1 string, a2 int) {})
	s, _ := vm.Compile("test.js", `x("zzz", 1)`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithIntVariadic0(b *testing.B) {
	vm := New()
	vm.Set("x", func(a ...int) {})
	s, _ := vm.Compile("test.js", `x()`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithIntVariadic1(b *testing.B) {
	vm := New()
	vm.Set("x", func(a ...int) {})
	s, _ := vm.Compile("test.js", `x(1)`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithIntVariadic3(b *testing.B) {
	vm := New()
	vm.Set("x", func(a ...int) {})
	s, _ := vm.Compile("test.js", `x(1, 2, 3)`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithIntVariadic10(b *testing.B) {
	vm := New()
	vm.Set("x", func(a ...int) {})
	s, _ := vm.Compile("test.js", `x(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithIntArray0(b *testing.B) {
	vm := New()
	vm.Set("x", func(a []int) {})
	s, _ := vm.Compile("test.js", `x([])`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithIntArray1(b *testing.B) {
	vm := New()
	vm.Set("x", func(a []int) {})
	s, _ := vm.Compile("test.js", `x([1])`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithIntArray3(b *testing.B) {
	vm := New()
	vm.Set("x", func(a []int) {})
	s, _ := vm.Compile("test.js", `x([1, 2, 3])`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithIntArray10(b *testing.B) {
	vm := New()
	vm.Set("x", func(a []int) {})
	s, _ := vm.Compile("test.js", `x([1, 2, 3, 4, 5, 6, 7, 8, 9, 10])`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithIntVariadicArray0(b *testing.B) {
	vm := New()
	vm.Set("x", func(a ...int) {})
	s, _ := vm.Compile("test.js", `x([])`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithIntVariadicArray1(b *testing.B) {
	vm := New()
	vm.Set("x", func(a ...int) {})
	s, _ := vm.Compile("test.js", `x([1])`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithIntVariadicArray3(b *testing.B) {
	vm := New()
	vm.Set("x", func(a ...int) {})
	s, _ := vm.Compile("test.js", `x([1, 2, 3])`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithIntVariadicArray10(b *testing.B) {
	vm := New()
	vm.Set("x", func(a ...int) {})
	s, _ := vm.Compile("test.js", `x([1, 2, 3, 4, 5, 6, 7, 8, 9, 10])`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithStringIntVariadic0(b *testing.B) {
	vm := New()
	vm.Set("x", func(a1 string, a2 ...int) {})
	s, _ := vm.Compile("test.js", `x("a")`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithStringIntVariadic1(b *testing.B) {
	vm := New()
	vm.Set("x", func(a1 string, a2 ...int) {})
	s, _ := vm.Compile("test.js", `x("a", 1)`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithStringIntVariadic3(b *testing.B) {
	vm := New()
	vm.Set("x", func(a1 string, a2 ...int) {})
	s, _ := vm.Compile("test.js", `x("a", 1, 2, 3)`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithStringIntVariadic10(b *testing.B) {
	vm := New()
	vm.Set("x", func(a1 string, a2 ...int) {})
	s, _ := vm.Compile("test.js", `x("a", 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithStringIntVariadicArray0(b *testing.B) {
	vm := New()
	vm.Set("x", func(a1 string, a2 ...int) {})
	s, _ := vm.Compile("test.js", `x("a", [])`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithStringIntVariadicArray1(b *testing.B) {
	vm := New()
	vm.Set("x", func(a1 string, a2 ...int) {})
	s, _ := vm.Compile("test.js", `x("a", [1])`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithStringIntVariadicArray3(b *testing.B) {
	vm := New()
	vm.Set("x", func(a1 string, a2 ...int) {})
	s, _ := vm.Compile("test.js", `x("a", [1, 2, 3])`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithStringIntVariadicArray10(b *testing.B) {
	vm := New()
	vm.Set("x", func(a1 string, a2 ...int) {})
	s, _ := vm.Compile("test.js", `x("a", [1, 2, 3, 4, 5, 6, 7, 8, 9, 10])`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithMap(b *testing.B) {
	vm := New()
	vm.Set("x", func(a map[string]string) {})
	s, _ := vm.Compile("test.js", `x({a: "b", c: "d"})`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithMapVariadic(b *testing.B) {
	vm := New()
	vm.Set("x", func(a ...map[string]string) {})
	s, _ := vm.Compile("test.js", `x({a: "b", c: "d"}, {w: "x", y: "z"})`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithMapVariadicArray(b *testing.B) {
	vm := New()
	vm.Set("x", func(a ...map[string]string) {})
	s, _ := vm.Compile("test.js", `x([{a: "b", c: "d"}, {w: "x", y: "z"}])`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithFunction(b *testing.B) {
	vm := New()
	vm.Set("x", func(a func()) {})
	s, _ := vm.Compile("test.js", `x(function() {})`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithFunctionInt(b *testing.B) {
	vm := New()
	vm.Set("x", func(a func(int)) {})
	s, _ := vm.Compile("test.js", `x(function(n) {})`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func BenchmarkNativeCallWithFunctionString(b *testing.B) {
	vm := New()
	vm.Set("x", func(a func(string)) {})
	s, _ := vm.Compile("test.js", `x(function(n) {})`)
	for i := 0; i < b.N; i++ {
		vm.Run(s)
	}
}

func TestNativeCallWithString(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a1 string) {
		if a1 != "zzz" {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x("zzz")`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithFloat32(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a1 float32) {
		if a1 != 1.1 {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x(1.1)`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithFloat64(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a1 float64) {
		if a1 != 1.1 {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x(1.1)`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithInt(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a1 int) {
		if a1 != 1 {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x(1)`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithUint(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a1 uint) {
		if a1 != 1 {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x(1)`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithInt8(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a1 int8) {
		if a1 != 1 {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x(1)`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithUint8(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a1 uint8) {
		if a1 != 1 {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x(1)`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithInt16(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a1 int16) {
		if a1 != 1 {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x(1)`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithUint16(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a1 uint16) {
		if a1 != 1 {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x(1)`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithInt32(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a1 int32) {
		if a1 != 1 {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x(1)`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithUint32(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a1 uint32) {
		if a1 != 1 {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x(1)`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithInt64(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a1 int64) {
		if a1 != 1 {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x(1)`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithUint64(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a1 uint64) {
		if a1 != 1 {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x(1)`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithStringInt(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a1 string, a2 int) {
		if a1 != "zzz" || a2 != 1 {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x("zzz", 1)`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithIntVariadic0(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a ...int) {
		if !reflect.DeepEqual(a, []int{}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x()`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithIntVariadic1(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a ...int) {
		if !reflect.DeepEqual(a, []int{1}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x(1)`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithIntVariadic3(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a ...int) {
		if !reflect.DeepEqual(a, []int{1, 2, 3}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x(1, 2, 3)`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithIntVariadic10(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a ...int) {
		if !reflect.DeepEqual(a, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithIntArray0(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a []int) {
		if !reflect.DeepEqual(a, []int{}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x([])`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithIntArray1(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a []int) {
		if !reflect.DeepEqual(a, []int{1}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x([1])`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithIntArray3(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a []int) {
		if !reflect.DeepEqual(a, []int{1, 2, 3}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x([1, 2, 3])`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithIntArray10(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a []int) {
		if !reflect.DeepEqual(a, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x([1, 2, 3, 4, 5, 6, 7, 8, 9, 10])`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithIntVariadicArray0(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a ...int) {
		if !reflect.DeepEqual(a, []int{}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x([])`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithIntVariadicArray1(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a ...int) {
		if !reflect.DeepEqual(a, []int{1}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x([1])`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithIntVariadicArray3(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a ...int) {
		if !reflect.DeepEqual(a, []int{1, 2, 3}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x([1, 2, 3])`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithIntVariadicArray10(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a ...int) {
		if !reflect.DeepEqual(a, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x([1, 2, 3, 4, 5, 6, 7, 8, 9, 10])`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithStringIntVariadic0(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a1 string, a2 ...int) {
		if a1 != "a" || !reflect.DeepEqual(a2, []int{}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x("a")`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithStringIntVariadic1(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a1 string, a2 ...int) {
		if a1 != "a" || !reflect.DeepEqual(a2, []int{1}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x("a", 1)`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithStringIntVariadic3(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a1 string, a2 ...int) {
		if a1 != "a" || !reflect.DeepEqual(a2, []int{1, 2, 3}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x("a", 1, 2, 3)`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithStringIntVariadic10(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a1 string, a2 ...int) {
		if a1 != "a" || !reflect.DeepEqual(a2, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x("a", 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithStringIntVariadicArray0(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a1 string, a2 ...int) {
		if a1 != "a" || !reflect.DeepEqual(a2, []int{}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x("a", [])`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithStringIntVariadicArray1(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a1 string, a2 ...int) {
		if a1 != "a" || !reflect.DeepEqual(a2, []int{1}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x("a", [1])`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithStringIntVariadicArray3(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a1 string, a2 ...int) {
		if a1 != "a" || !reflect.DeepEqual(a2, []int{1, 2, 3}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x("a", [1, 2, 3])`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithStringIntVariadicArray10(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a1 string, a2 ...int) {
		if a1 != "a" || !reflect.DeepEqual(a2, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x("a", [1, 2, 3, 4, 5, 6, 7, 8, 9, 10])`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithMap(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a map[string]string) {
		if !reflect.DeepEqual(a, map[string]string{"a": "b", "c": "d"}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x({a: "b", c: "d"})`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithMapVariadic(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a ...map[string]string) {
		if !reflect.DeepEqual(a, []map[string]string{{"a": "b", "c": "d"}, {"w": "x", "y": "z"}}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x({a: "b", c: "d"}, {w: "x", y: "z"})`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithMapVariadicArray(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(a ...map[string]string) {
		if !reflect.DeepEqual(a, []map[string]string{{"a": "b", "c": "d"}, {"w": "x", "y": "z"}}) {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x([{a: "b", c: "d"}, {w: "x", y: "z"}])`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithFunctionVoidBool(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(fn func() bool) {
		if !fn() {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x(function() { return true; })`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithFunctionIntInt(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(fn func(int) int) {
		if fn(5) != 5 {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x(function(n) { return n; })`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallWithFunctionStringString(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", func(fn func(string) string) {
		if fn("zzz") != "zzz" {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `x(function(n) { return n; })`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

type testNativeCallWithStruct struct {
	Prefix string
}

type testNativeCallWithStructArg struct {
	Text string
}

func (t testNativeCallWithStruct) MakeStruct(s string) testNativeCallWithStructArg {
	return testNativeCallWithStructArg{Text: s}
}

func (t testNativeCallWithStruct) MakeStructPointer(s string) *testNativeCallWithStructArg {
	return &testNativeCallWithStructArg{Text: s}
}

func (t testNativeCallWithStruct) CallWithStruct(a testNativeCallWithStructArg) string {
	return t.Prefix + a.Text
}

func (t *testNativeCallWithStruct) CallPointerWithStruct(a testNativeCallWithStructArg) string {
	return t.Prefix + a.Text
}

func (t testNativeCallWithStruct) CallWithStructPointer(a *testNativeCallWithStructArg) string {
	return t.Prefix + a.Text
}

func (t *testNativeCallWithStruct) CallPointerWithStructPointer(a *testNativeCallWithStructArg) string {
	return t.Prefix + a.Text
}

func TestNativeCallMethodWithStruct(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", testNativeCallWithStruct{Prefix: "a"})

	vm.Set("t", func(s string) {
		if s != "ab" {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `t(x.CallWithStruct(x.MakeStruct("b")))`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallPointerMethodWithStruct(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", &testNativeCallWithStruct{Prefix: "a"})

	vm.Set("t", func(s string) {
		if s != "ab" {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `t(x.CallPointerWithStruct(x.MakeStruct("b")))`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallMethodWithStructPointer(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", testNativeCallWithStruct{Prefix: "a"})

	vm.Set("t", func(s string) {
		if s != "ab" {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `t(x.CallWithStructPointer(x.MakeStructPointer("b")))`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallPointerMethodWithStructPointer(t *testing.T) {
	vm := New()

	called := false

	vm.Set("x", &testNativeCallWithStruct{Prefix: "a"})

	vm.Set("t", func(s string) {
		if s != "ab" {
			t.Fail()
		}

		called = true
	})

	s, _ := vm.Compile("test.js", `t(x.CallPointerWithStructPointer(x.MakeStructPointer("b")))`)

	if _, err := vm.Run(s); err != nil {
		t.Logf("err should have been nil; was %s\n", err.Error())
		t.Fail()
	}

	if !called {
		t.Fail()
	}
}

func TestNativeCallNilInterfaceArg(t *testing.T) {
	vm := New()
	vm.Set("f1", func(v interface{}) {})
	vm.Call("f1", nil, nil)
}
