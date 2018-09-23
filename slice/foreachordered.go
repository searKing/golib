package slice

import "github.com/searKing/golib/util/object"

// ForEachOrderedFunc Performs an action for each element of this slice.
// <p>This operation processes the elements one at a time, in encounter
// order if one exists.  Performing the action for one element
// performing the action for subsequent elements, but for any given element,
// the action may be performed in whatever thread the library chooses.
func ForEachOrderedFunc(s interface{}, f func(interface{}), ifStringAsRune ...bool) {
	forEachOrderedFunc(Of(s, ifStringAsRune...), f)
}

// forEachOrderedFunc is the same as ForEachOrderedFunc
func forEachOrderedFunc(s []interface{}, f func(interface{})) {
	object.RequireNonNil(s, "forEachOrderedFunc called on nil slice")
	object.RequireNonNil(s, "forEachOrderedFunc called on nil callfn")

	for _, r := range s {
		f(r)
	}
	return
}
