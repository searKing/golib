package slice

import "github.com/searKing/golib/util/object"

// TakeWhileFunc returns whether any elements of this stream match the provided
// predicate.  May not evaluate the predicate on all elements if not
// necessary for determining the result.  If the stream is empty then
// {@code false} is returned and the predicate is not evaluated.
func AnyMatchFunc(s interface{}, f func(interface{}) bool, ifStringAsRune ...bool) bool {
	return anyMatchFunc(Of(s, ifStringAsRune...), f, true)
}

// anyMatchFunc is the same as AnyMatchFunc.
func anyMatchFunc(s []interface{}, f func(interface{}) bool, truth bool) bool {
	object.RequireNonNil(s, "anyMatchFunc called on nil slice")
	object.RequireNonNil(f, "anyMatchFunc called on nil callfn")

	for _, r := range s {
		if f(r) == truth {
			return true
		}
	}
	return false
}
