package slice

import "strings"

func sd() {
	strings.IndexFunc()
}

// IndexFunc returns the index into s of the first Unicode
// code point satisfying f(c), or -1 if none do.
func MapFunc(s []interface{}, f func(interface{}) bool) []interface{} {
	return mapFunc(s, f, true)
}

// indexFunc is the same as IndexFunc except that if
// truth==false, the sense of the predicate function is
// inverted.
func mapFunc(s []interface{}, f func(interface{}) bool, truth bool) []interface{} {
	var sMapped = []interface{}{}
	for _, r := range s {
		if f(r) == truth {
			sMapped = append(sMapped, r)
		}
	}
	return sMapped
}
