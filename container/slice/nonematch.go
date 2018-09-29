package slice

// NoneMatchFunc returns whether no elements of this stream match the provided predicate.
// May not evaluate the predicate on all elements if not necessary for
// determining the result.  If the stream is empty then {@code true} is
// returned and the predicate is not evaluated.
func NoneMatchFunc(s interface{}, f func(interface{}) bool, ifStringAsRune ...bool) bool {
	return noneMatchFunc(Of(s, ifStringAsRune...), f, true)
}

// noneMatchFunc is the same as NoneMatchFunc.
func noneMatchFunc(s []interface{}, f func(interface{}) bool, truth bool) bool {
	return !anyMatchFunc(s, f, truth)
}
