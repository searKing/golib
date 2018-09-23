package slice

import "github.com/searKing/golib/util/object"

// FilterFunc returns a slice consisting of the elements of this slice that match
//the given predicate.
func FilterFunc(s interface{}, f func(interface{}) bool, ifStringAsRune ...bool) interface{} {
	return normalizeSlice(filterFunc(Of(s, ifStringAsRune...), f, true), s, ifStringAsRune...)

}

// filterFunc is the same as FilterFunc except that if
// truth==false, the sense of the predicate function is
// inverted.
func filterFunc(s []interface{}, f func(interface{}) bool, truth bool) []interface{} {
	object.RequireNonNil(s, "filterFunc called on nil slice")
	object.RequireNonNil(s, "filterFunc called on nil callfn")

	var sFiltered = []interface{}{}
	for _, r := range s {
		if f(r) == truth {
			sFiltered = append(sFiltered, r)
		}
	}
	return sFiltered
}
