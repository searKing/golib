package default_

import (
	"reflect"
	"sync"
)

// An encodeState encodes JSON into a bytes.Buffer.
type convertState struct {
}

func (_ *convertState) Reset() {
	return
}

var convertStatePool sync.Pool

type convertFunc func(e *convertState, v reflect.Value, opts convOpts)

func newConvertState() *convertState {
	if v := convertStatePool.Get(); v != nil {
		e := v.(*convertState)
		e.Reset()
		return e
	}
	return new(convertState)
}

// defaultError is an error wrapper type for internal use only.
// Panics with errors are wrapped in defaultError so that the top-level recover
// can distinguish intentional panics from this package.
type defaultError struct{ error }

func (e *convertState) convert(v interface{}, opts convOpts) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if je, ok := r.(defaultError); ok {
				err = je.error
			} else {
				panic(r)
			}
		}
	}()
	e.reflectValue(reflect.ValueOf(v), opts)
	return nil
}

// error aborts the encoding by panicking with err wrapped in defaultError.
func (e *convertState) error(err error) {
	panic(defaultError{err})
}
func (e *convertState) reflectValue(v reflect.Value, opts convOpts) {
	valueConverter(v)(e, v, opts)
}
