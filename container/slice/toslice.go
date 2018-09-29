package slice

// ToSliceFunc returns an array containing the elements of this stream.
func ToSliceFunc(s interface{}, ifStringAsRune ...bool) interface{} {
	return toSliceFunc(Of(s, ifStringAsRune...))
}

// toSliceFunc is the same as ToSliceFunc
func toSliceFunc(s []interface{}) []interface{} {
	return s
}
