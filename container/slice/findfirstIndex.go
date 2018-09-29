package slice

import (
	"github.com/searKing/golib/util/object"
)

// FindFirstFunc returns an {@link Optional} describing the first index of this stream,
// or an empty {@code Optional} if the stream is empty.  If the stream has
// no encounter order, then any element may be returned.
func FindFirstIndexFunc(s interface{}, f func(interface{}) bool, ifStringAsRune ...bool) int {
	return findFirstIndexFunc(Of(s, ifStringAsRune...), f, true)
}

// findFirstFunc is the same as FindFirstFunc.
func findFirstIndexFunc(s []interface{}, f func(interface{}) bool, truth bool) int {
	object.RequireNonNil(s, "findFirstIndexFunc called on nil slice")
	object.RequireNonNil(f, "findFirstIndexFunc called on nil callfn")

	for idx, r := range s {
		if f(r) == truth {
			return idx
		}
	}
	return -1
}
