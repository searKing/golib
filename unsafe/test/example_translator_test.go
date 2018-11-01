package unsafetest_test

import (
	"github.com/searKing/golib/unsafe/test"
)

// The actual test functions are in non-_test.go files
// so that they can use cgo (import "C").
// These wrappers are here for gotest to find.
func ExampleGoStringArray() { unsafetest.ExampleGoStringArray() }
func ExampleCStringArray()  { unsafetest.ExampleCStringArray() }
