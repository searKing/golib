package slice

import (
	"github.com/searKing/golib/util/object"
	"reflect"
)

// ConcatFunc creates a lazily concatenated stream whose elements are all the
// elements of the first stream followed by all the elements of the
// second stream.  The resulting stream is ordered if both
// of the input streams are ordered, and parallel if either of the input
// streams is parallel.  When the resulting stream is closed, the close
// handlers for both input streams are invoked.
func ConcatFunc(s1, s2 interface{}) interface{} {
	return concatFunc(s1, s2)
}

// concatFunc is the same as ConcatFunc
func concatFunc(s1, s2 interface{}) interface{} {
	object.RequireNonNil(s1, "concatFunc called on nil slice")
	object.RequireNonNil(s2, "concatFunc called on nil callfn")
	typ1 := reflect.ValueOf(s1).Kind()
	typ2 := reflect.ValueOf(s2).Kind()
	object.RequireEqual(typ1, typ2)
	if typ1 == reflect.String {
		return s1.(string) + s2.(string)
	}

	var sConcated = []interface{}{}
	for _, r := range Of(s1) {
		sConcated = append(sConcated, r)
	}
	for _, r := range Of(s2) {
		sConcated = append(sConcated, r)
	}
	return sConcated
}
