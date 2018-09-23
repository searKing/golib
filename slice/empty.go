package slice

// EmptyFunc returns an empty sequential {@code slice}.
func EmptyFunc(s interface{}, ifStringAsRune ...bool) interface{} {
	return normalizeSlice(emptyFunc(), s, ifStringAsRune...)
}

// emptyFunc is the same as EmptyFunc
func emptyFunc() []interface{} {
	return []interface{}{}
}
