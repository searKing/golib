package slice

import "github.com/searKing/golib/util/object"

//Returns a slice consisting of the results of applying the given
//function to the elements of this slice.
func MapFunc(s []interface{}, f func(interface{}) interface{}) []interface{} {
	return mapFunc(s, f)
}

// mapFunc is the same as MapFunc
func mapFunc(s []interface{}, f func(interface{}) interface{}) []interface{} {
	object.RequireNonNil(s, "mapFunc called on nil slice")
	object.RequireNonNil(s, "mapFunc called on nil callfn")

	var sMapped = []interface{}{}
	for _, r := range s {
		sMapped = append(sMapped, f(r))
	}
	return sMapped
}
