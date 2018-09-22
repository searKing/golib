// +build !debug

package error

// Assert panics ignored.
func Assert(cond bool, a ...interface{}) {}

// Assertln panics ignored.
func Assertln(cond bool, a ...interface{}) {}

// Assertf panics ignored.
func Assertf(cond bool, format string, a ...interface{}) {}
