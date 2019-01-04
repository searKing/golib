package default_

import "reflect"

type ConverterError struct {
	Type reflect.Type
	Err  error
}

func (e *ConverterError) Error() string {
	return "default: error calling ConvertDefault for type " + e.Type.String() + ": " + e.Err.Error()
}

// An UnsupportedTypeError is returned by Marshal when attempting
// to convert an unsupported value type.
type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "default: unsupported type: " + e.Type.String()
}

type UnsupportedValueError struct {
	Value reflect.Value
	Str   string
}

func (e *UnsupportedValueError) Error() string {
	return "default: unsupported value: " + e.Str
}
