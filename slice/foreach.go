package slice

import "github.com/searKing/golib/util/object"

//Performs an action for each element of this slice.
func ForEachFunc(s interface{}, f func(interface{}), ifStringAsRune ...bool) interface{} {
	return normalizeSlice(mapFunc(Of(s, ifStringAsRune...), f))
}

// forEachFunc is the same as ForEachFunc
func forEachFunc(s []interface{}, f func(interface{})) {
	object.RequireNonNil(s, "mapFunc called on nil slice")
	object.RequireNonNil(s, "mapFunc called on nil callfn")

	for _, r := range s {
		f(r)
	}
	return
}
