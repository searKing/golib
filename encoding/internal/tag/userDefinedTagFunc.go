package tag

import "reflect"

// Convert v
func userDefinedTagFunc(e *tagState, v reflect.Value, _ tagOpts) {
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return
	}
	m, ok := v.Interface().(Tagger)
	if !ok {
		return
	}
	err := m.TagDefault()
	if err != nil {
		e.error(&TaggerError{v.Type(), err})
	}
}

// Convert &v
func addrUserDefinedTagFunc(e *tagState, v reflect.Value, _ tagOpts) {
	va := v.Addr()
	if va.IsNil() {
		return
	}
	m := va.Interface().(Tagger)
	err := m.TagDefault()

	if err != nil {
		e.error(&TaggerError{v.Type(), err})
	}
}
