package slice

import (
	"github.com/searKing/golib/util/object"
	"reflect"
	"sort"
)

// SortedFunc returns a stream consisting of the distinct elements (according to
// {@link Object#equals(Object)}) of this stream.
// s: Accept Array、Slice、String(as []byte if ifStringAsRune else []rune)
func SortedFunc(s interface{}, f func(interface{}, interface{}) int, ifStringAsRune ...bool) interface{} {
	sorted := sortedFunc(Of(s, ifStringAsRune...), f)
	if kind := reflect.ValueOf(s).Kind(); kind != reflect.String {
		return sorted
	}

	// AS []rune
	if isAsRune(ifStringAsRune...) {
		bs := make([]rune, 0, len(sorted))
		for _, s := range sorted {
			bs = append(bs, s.(rune))
		}
		return string(bs)
	}
	// AS []byte
	bs := make([]byte, len(sorted))
	for _, s := range sorted {
		bs = append(bs, s.(byte))
	}
	return string(bs)
}

// sortedFunc is the same as SortedFunc except that if
// truth==false, the sense of the predicate function is
// inverted.
func sortedFunc(s []interface{}, f func(interface{}, interface{}) int) []interface{} {
	object.RequireNonNil(s, "distinctFunc called on nil slice")
	object.RequireNonNil(f, "distinctFunc called on nil callfn")

	less := func(i, j int) bool {
		if f(s[i], s[j]) < 0 {
			return true
		}
		return false
	}
	sort.Slice(s, less)
	return s
}
