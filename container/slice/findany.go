package slice

// FindAnyFunc returns an {@link Optional} describing some element of the stream, or an
// empty {@code Optional} if the stream is empty.
func FindAnyFunc(s interface{}, f func(interface{}) bool) interface{} {
	return normalizeElem(findAnyFunc(Of(s), f, true), s)
}

// findAnyFunc is the same as FindAnyFunc.
func findAnyFunc(s []interface{}, f func(interface{}) bool, truth bool) interface{} {
	idx := findAnyIndexFunc(s, f, truth)
	if idx == -1 {
		return nil
	}
	return s[idx]
}
