package slice

import (
	"github.com/searKing/golib/util/object"
	"github.com/searKing/golib/util/optional"
)

// ReduceFunc calls a defined callback function on each element of an array, and returns an array that contains the results.
func ReduceFunc(s interface{}, f func(left, right interface{}) interface{}, ifStringAsRune ...bool) interface{} {
	return normalizeElem(reduceFunc(Of(s, ifStringAsRune...), f), s)

}

// reduceFunc is the same as ReduceFunc
func reduceFunc(s []interface{}, f func(left, right interface{}) interface{}, identity ...interface{}) interface{} {
	object.RequireNonNil(s, "reduceFunc called on nil slice")
	object.RequireNonNil(s, "reduceFunc called on nil callfn")

	var foundAny bool
	var result interface{}

	if (identity != nil || len(identity) != 0) {
		foundAny = true;
		result = identity;
	}
	for _, r := range s {
		if (!foundAny) {
			foundAny = true;
			result = r;
		} else {
			result = f(result, r);
		}
	}
	if foundAny {
		return optional.Of(result).Get()
	}
	return optional.Empty().Get()
}
