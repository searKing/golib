package slice

import (
	"github.com/searKing/golib/util/object"
)

// CountFunc returns the maximum element of this stream according to the provided.
func CountFunc(s interface{}, ifStringAsRune ...bool) int {
	return countFunc(Of(s, ifStringAsRune...))

}

// countFunc is the same as CountFunc
func countFunc(s []interface{}) int {
	object.RequireNonNil(s, "countFunc called on nil slice")
	return len(s)
}
