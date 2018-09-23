package slice

import "github.com/searKing/golib/util/object"

// DistinctFunc returns a slice consisting of the distinct elements (according to
// {@link Object#equals(Object)}) of this slice.
func DistinctFunc(s []interface{}, f func(interface{}, interface{}) int) []interface{} {
	return distinctFunc(s, f)
}

// distinctFunc is the same as DistinctFunc except that if
//// truth==false, the sense of the predicate function is
//// inverted.
func distinctFunc(s []interface{}, f func(interface{}, interface{}) int) []interface{} {
	object.RequireNonNil(s, "distinctFunc called on nil slice")
	object.RequireNonNil(s, "distinctFunc called on nil callfn")

	sDistinctMap := map[interface{}]struct{}{}
	var sDistincted = []interface{}{}
	for _, r := range s {
		if _, ok := sDistinctMap[r]; ok {
			continue
		}
		sDistinctMap[r] = struct{}{}
		sDistincted = append(sDistincted, r)
	}
	return sDistincted
}
