package slice

import (
	"github.com/searKing/golib/util/object"
)

// MinFunc returns the minimum element of this stream according to the provided.
func MinFunc(s interface{}, f func(interface{}, interface{}) int, ifStringAsRune ...bool) interface{} {
	return normalizeElem(minFunc(Of(s, ifStringAsRune...), f), s)

}

// minFunc is the same as MinFunc
func minFunc(s []interface{}, f func(interface{}, interface{}) int, ifStringAsRune ...bool) interface{} {
	object.RequireNonNil(s, "minFunc called on nil slice")
	object.RequireNonNil(s, "minFunc called on nil callfn")

	return ReduceFunc(s, func(left, right interface{}) interface{} {
		if f(left, right) < 0 {
			return left
		}
		return right
	}, ifStringAsRune...)
}
