package object

import (
	"errors"
	"strings"
)

type ErrorNilPointer error

var (
	errorNilPointer = errors.New("nil pointer")
)

type Supplier interface {
	Get() interface{}
}

// RequireNonNil checks that the specified object reference is not {@code nil}. This
// method is designed primarily for doing parameter validation in methods
// and constructors
func RequireNonNil(obj interface{}, msg ...string) interface{} {
	if msg == nil {
		msg = []string{"nil pointer"}
	}
	if obj == nil {
		panic(ErrorNilPointer(errors.New(strings.Join(msg, ""))))
	}
	return obj
}

// IsNil returns {@code true} if the provided reference is {@code nil} otherwise
// returns {@code false}.
func IsNil(obj interface{}) bool {
	return obj == nil
}

// IsNil returns {@code true} if the provided reference is non-{@code nil} otherwise
// returns {@code false}.
func NoneNil(obj interface{}) bool {
	return obj != nil
}

// RequireNonNullElse returns the first argument if it is non-{@code nil} and
// otherwise returns the non-{@code nil} second argument.
func RequireNonNullElse(obj, defaultObj interface{}) interface{} {
	if obj != nil {
		return obj
	}
	return RequireNonNil(defaultObj, "defaultObj")
}

// RequireNonNullElseGet returns the first argument if it is non-{@code nil} and
// returns the non-{@code nil} value of {@code supplier.Get()}.
func RequireNonNullElseGet(obj interface{}, supplier Supplier) interface{} {
	if obj != nil {
		return obj
	}
	return RequireNonNil(RequireNonNil(supplier, "supplier").(Supplier).Get(), "supplier.Get()")
}
