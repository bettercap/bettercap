package otto

import "testing"

type GoSliceTest []int

func (s GoSliceTest) Sum() int {
	sum := 0
	for _, v := range s {
		sum += v
	}
	return sum
}

func TestGoSlice(t *testing.T) {
	tt(t, func() {
		test, vm := test()
		vm.Set("TestSlice", GoSliceTest{1, 2, 3})
		is(test(`TestSlice.length`).export(), 3)
		is(test(`TestSlice[1]`).export(), 2)
		is(test(`TestSlice.Sum()`).export(), 6)
	})
}
