package slice

import "github.com/searKing/golib/util/object"

// FindFirstFunc returns an {@link Optional} describing the first element of this stream,
// or an empty {@code Optional} if the stream is empty.  If the stream has
// no encounter order, then any element may be returned.
func FindFirstFunc(s interface{}, f func(interface{}) bool, ifStringAsRune ...bool) interface{} {
	return normalizeElem(findFirstFunc(Of(s, ifStringAsRune...), f, true), s, ifStringAsRune...)
}

// findFirstFunc is the same as FindFirstFunc.
func findFirstFunc(s []interface{}, f func(interface{}) bool, truth bool) interface{} {
	object.RequireNonNil(s, "findFirstFunc called on nil slice")
	object.RequireNonNil(f, "findFirstFunc called on nil callfn")

	for _, r := range s {
		if f(r) == truth {
			return r
		}
	}
	return nil
}
