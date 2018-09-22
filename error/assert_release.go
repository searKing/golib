// +build !debug

package error

// Assert panics if cond is false. Truef formats the panic message using the
// default formats for its operands.
func Assert(cond bool, a ...interface{}) {}

// Assertf panics if cond is false. Truef formats the panic message according to a
// format specifier.
func Assertf(cond bool, format string, a ...interface{}) {}
