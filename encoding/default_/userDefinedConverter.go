package default_

import "reflect"

// Convert v
func userDefinedConverter(e *convertState, v reflect.Value, _ convOpts) {
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return
	}
	m, ok := v.Interface().(Converter)
	if !ok {
		return
	}
	err := m.ConvertDefault()
	if err != nil {
		e.error(&ConverterError{v.Type(), err})
	}
}

// Convert &v
func addrUserDefinedConverter(e *convertState, v reflect.Value, _ convOpts) {
	va := v.Addr()
	if va.IsNil() {
		return
	}
	m := va.Interface().(Converter)
	err := m.ConvertDefault()

	if err != nil {
		e.error(&ConverterError{v.Type(), err})
	}
}
