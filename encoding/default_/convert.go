package default_

import (
	"reflect"
)

type convOpts struct{}

func Convert(v interface{}) error {
	e := newConvertState()
	err := e.convert(v, convOpts{})
	if err != nil {
		return err
	}

	e.Reset()
	convertStatePool.Put(e)
	return nil
}

// Marshaler is the interface implemented by types that
// can marshal themselves into valid JSON.
type Converter interface {
	ConvertDefault() error
}

func invalidValueConverter(e *convertState, v reflect.Value, _ convOpts) {
}
func unsupportedTypeConverter(e *convertState, v reflect.Value, _ convOpts) {
	e.error(&UnsupportedTypeError{v.Type()})
}
