package otto

import (
	"testing"
)

func TestCloneGetterSetter(t *testing.T) {
	vm := New()

	vm.Run(`var x = Object.create(null, {
    x: {
      get: function() {},
      set: function() {},
    },
  })`)

	vm.Copy()
}
