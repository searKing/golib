package default_

import (
	"reflect"
)

var converterCache converterCacheMap // map[reflect.Type]convertFunc

var converterType = reflect.TypeOf(new(Converter)).Elem()

func valueConverter(v reflect.Value) convertFunc {
	if !v.IsValid() {
		return invalidValueConverter
	}
	return typeConverter(v.Type())
}
