package default_

import "reflect"

func unsupportedTypeConverter(e *convertState, v reflect.Value, _ convOpts) {
	e.error(&UnsupportedTypeError{v.Type()})
}
