package slice



// IndexFunc returns the index into s of the first Unicode
// code point satisfying f(c), or -1 if none do.
func ReduceFunc(s []interface{}, f func(left, right interface{}) interface{}) interface{} {
	return reduceFunc(s, f)
}

// indexFunc is the same as IndexFunc except that if
// truth==false, the sense of the predicate function is
// inverted.

// Calls a defined callback function on each element of an array, and returns an array that contains the results.
// @param f A function that accepts up to three arguments. The map method calls the f function one time for each element in the array.
// @param thisArg An object to which the this keyword can refer in the callbackfn function. If thisArg is omitted, undefined is used as the this value.

func reduceFunc(s []interface{}, f func(left, right interface{}) interface{}) Optional {
	if s == nil {
		panic("reduce called on nil slice")
	}
	if f == nil {
		panic("reduce called on nil callfn")
	}
	var foundAny bool
	var result interface{}
	for _, r := range s {
		if (!foundAny) {
			foundAny = true;
			result = r;
		} else {
			result = f(result, r);
		}
	}
	if foundAny {
		return Optional{
			value: result,
		}
	}
	return OptionalEmpty
}
