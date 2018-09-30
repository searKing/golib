package boolean

import (
	"github.com/searKing/golib/container/slice"
	"github.com/searKing/golib/util/object"
)

// https://en.wikipedia.org/wiki/Boolean_operation

func xor(a, b bool) bool {
	return a != b
}
func xnor(a, b bool) bool {
	return !xor(a, b)
}
func or(a, b bool) bool {
	return a || b
}
func and(a, b bool) bool {
	return a && b
}

func RelationFunc(a, b interface{}, f func(a, b interface{}) interface{}, c ...interface{}) interface{} {
	object.RequireNonNil(f)
	if c == nil || len(c) == 0 {
		return f(a, b)
	}
	return RelationFunc(f(a, b), c[0], f, c[1:]...)
}
func BoolFunc(a bool, b bool, f func(a, b bool) bool, c ...bool) bool {

	return RelationFunc(a, b, func(a, b interface{}) interface{} {
		return f(a.(bool), b.(bool))
	}, slice.Of(c)...).(bool)
	object.RequireNonNil(f)
	if c == nil || len(c) == 0 {
		return f(a, b)
	}
	return BoolFunc(f(a, b), c[0], f, c[1:]...)
}

func XOR(a bool, b bool, c ...bool) bool {
	return BoolFunc(a, b, xor, c...)
}
func XNOR(a bool, b bool, c ...bool) bool {
	return BoolFunc(a, b, xnor, c...)
}

func OR(a bool, b bool, c ...bool) bool {
	return BoolFunc(a, b, or, c...)
}

func AND(a bool, b bool, c ...bool) bool {
	return BoolFunc(a, b, and, c...)
}
