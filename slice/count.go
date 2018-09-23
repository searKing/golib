package slice

import (
	"github.com/searKing/golib/util/object"
)

// CountFunc returns the maximum element of this stream according to the provided.
func CountFunc(s []interface{}, f func(interface{}, interface{}) int, ifStringAsRune ...bool) interface{} {
	return minFunc(Of(s, ifStringAsRune...), f)

}

// countFunc is the same as CountFunc
func countFunc(s []interface{}) int {
	object.RequireNonNil(s, "countFunc called on nil slice")
	return len(s)
}
