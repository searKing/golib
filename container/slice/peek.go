package slice

import "github.com/searKing/golib/util/object"

// PeekFunc returns a slice consisting of the elements of this slice, additionally
// performing the provided action on each element as elements are consumed
// from the resulting slice.
func PeekFunc(s interface{}, f func(interface{})) interface{} {
	return normalizeSlice(peekFunc(Of(s), f), s)

}

// peekFunc is the same as PeekFunc.
func peekFunc(s []interface{}, f func(interface{})) []interface{} {
	object.RequireNonNil(s, "peekFunc called on nil slice")
	object.RequireNonNil(f, "peekFunc called on nil callfn")

	for _, r := range s {
		f(r)
	}
	return s
}
