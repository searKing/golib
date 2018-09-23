package slice

import "github.com/searKing/golib/util/object"

// LimitFunc returns a slice consisting of the distinct elements (according to
// {@link Object#equals(Object)}) of this slice.
func LimitFunc(s interface{}, maxSize int, ifStringAsRune ...bool) interface{} {
	return normalizeSlice(limitFunc(Of(s, ifStringAsRune...), maxSize))
}

// limitFunc is the same as DistinctFunc except that if
//// truth==false, the sense of the predicate function is
//// inverted.
func limitFunc(s []interface{}, maxSize int) []interface{} {
	object.RequireNonNil(s, "distinctFunc called on nil slice")
	m := len(s)
	if m > maxSize {
		m = maxSize
	}
	return s[:m]
}
