// +build debug

package error

import "fmt"

// Assert panics if cond is false.
func Assert(cond bool, a ...interface{}) {
	Assertf(cond, fmt.Sprint(a...))
}

// Assertln panics if cond is false.
func Assertln(cond bool, a ...interface{}) {
	Assertf(cond, fmt.Sprintln(a...))
}
// Assertf panics if cond is false.
func Assertf(cond bool, format string, a ...interface{}) {
	if !cond {
		panic(fmt.Sprintf(format, a...))
	}
}
