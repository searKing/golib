package slice

// EmptyFunc returns an empty sequential {@code slice}.
func EmptyFunc() interface{} {
	return emptyFunc()
}

// emptyFunc is the same as EmptyFunc
func emptyFunc() []interface{} {
	return []interface{}{}
}
