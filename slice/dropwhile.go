package slice

import "github.com/searKing/golib/util/object"

// DropWhileFunc returns, if this slice is ordered, a slice consisting of the remaining
// elements of this slice after dropping the longest prefix of elements
// that match the given predicate.  Otherwise returns, if this slice is
// unordered, a slice consisting of the remaining elements of this slice
// after dropping a subset of elements that match the given predicate.
func DropWhileFunc(s interface{}, f func(interface{}) bool, ifStringAsRune ...bool) interface{} {
	return normalizeSlice(dropWhileFunc(Of(s, ifStringAsRune...), f, true))
}

// dropWhileFunc is the same as DropWhileFunc.
func dropWhileFunc(s []interface{}, f func(interface{}) bool, truth bool) []interface{} {
	object.RequireNonNil(s, "takeWhileFunc called on nil slice")
	object.RequireNonNil(f, "takeWhileFunc called on nil callfn")

	var sTaken = []interface{}{}
	dropFound := false
	for _, r := range s {
		if !dropFound && f(r) == truth {
			continue
		}
		dropFound = true
		sTaken = append(sTaken, r)
	}
	return sTaken
}
