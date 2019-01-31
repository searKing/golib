package access

import "fmt"

type BadStringError struct {
	What string
	Str  string
}

func (e *BadStringError) Error() string { return fmt.Sprintf("%s %q", e.What, e.Str) }
