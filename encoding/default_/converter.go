package default_

import (
	"reflect"
)

type convertFunc func(e *convertState, v reflect.Value, opts convOpts)

func valueConverter(v reflect.Value) convertFunc {
	if !v.IsValid() {
		return invalidValueConverter
	}
	return typeConverter(v.Type())
}
