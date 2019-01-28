package access

import "fmt"

type BadStringError struct {
	what string
	str  string
}

func (e *BadStringError) Error() string { return fmt.Sprintf("%s %q", e.what, e.str) }
