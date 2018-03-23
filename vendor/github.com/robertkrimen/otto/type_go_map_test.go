package otto

import (
	"fmt"
	"sort"
	"testing"
)

type GoMapTest map[string]int

func (s GoMapTest) Join() string {
	joinedStr := ""

	// Ordering the map takes some effort
	// because map iterators in golang are unordered by definition.
	// So we need to extract keys, sort them, and then generate K/V pairs
	// All of this is meant to ensure that the test is predictable.
	keys := make([]string, len(s))
	i := 0
	for key, _ := range s {
		keys[i] = key
		i++
	}

	sort.Strings(keys)

	for _, key := range keys {
		joinedStr += key + ": " + fmt.Sprintf("%d", s[key]) + " "
	}
	return joinedStr
}

func TestGoMap(t *testing.T) {
	tt(t, func() {
		test, vm := test()
		vm.Set("TestMap", GoMapTest{"one": 1, "two": 2, "three": 3})
		is(test(`TestMap["one"]`).export(), 1)
		is(test(`TestMap.Join()`).export(), "one: 1 three: 3 two: 2 ")
	})
}
