package slice

import (
	"reflect"
)

// Of returns a slice consisting of the elements.
// obj: Accept Array、Slice、String(as []byte if ifStringAsRune else []rune)
func Of(obj interface{}, ifStringAsRune ...bool) []interface{} {
	return of(obj, isAsRune(ifStringAsRune...))
}

//of is the same as Of
func of(obj interface{}, ifStringAsRune bool) []interface{} {
	switch kind := reflect.ValueOf(obj).Kind(); kind {
	default:
		panic(&reflect.ValueError{"reflect.Value.Slice", kind})
	case reflect.Array, reflect.Slice:
	case reflect.String:
		if ifStringAsRune {
			out := []interface{}{}
			in := obj.(string)
			for _, r := range in {
				out = append(out, r)
			}
			return out
		}
	}

	out := []interface{}{}
	v := reflect.ValueOf(obj)
	for i := 0; i < v.Len(); i++ {
		out = append(out, v.Slice(i, i+1).Index(0).Interface())
	}
	return out
}
