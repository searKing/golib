package slice

import "github.com/searKing/golib/util/object"

// SkipFunc returns a slice consisting of the remaining elements of this slice
// after discarding the first {@code n} elements of the slice.
// If this slice contains fewer than {@code n} elements then an
// empty slice will be returned.
func SkipFunc(s interface{}, n int) interface{} {
	return normalizeSlice(skipFunc(Of(s), n), s)
}

// skipFunc is the same as SkipFunc.
func skipFunc(s []interface{}, n int) []interface{} {
	object.RequireNonNil(s, "skipFunc called on nil slice")
	m := len(s)
	if m > n {
		m = n
	}
	return s[m:]
}
