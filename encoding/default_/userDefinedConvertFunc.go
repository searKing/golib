package default_

import "reflect"

// Convert v
func userDefinedConvertFunc( /*userDefinedInterfaceType reflect.Type,*/ v reflect.Value, tag reflect.StructTag) error {
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil
	}
	//for idx := 0; idx < v.NumMethod(); idx++{
	//	if v.Method(idx).Type().Implements(userDefinedInterfaceType) {
	//
	//	}
	//}
	//reflect.TypeOf(new(Converter)).Elem().Method(0).Func.Call([]reflect.Value{v, reflect.ValueOf(tag)})
	m, ok := v.Interface().(Converter)
	if !ok {
		return nil
	}
	return m.ConvertDefault(v, tag)
}

// Convert &v
func addrUserDefinedConvertFunc(v reflect.Value, tag reflect.StructTag) error {
	va := v.Addr()
	if va.IsNil() {
		return nil
	}
	m := va.Interface().(Converter)
	return m.ConvertDefault(v, tag)
}

// newTypeConverter constructs an convertorFunc for a type.
// The returned encoder only checks CanAddr when allowAddr is true.
func newTypeConverter(convFn convertFunc, t reflect.Type, allowAddr bool) convertFunc {
	// Handle UserDefined Case
	// Convert v
	if t.Implements(converterType) {
		return userDefinedConvertFunc
	}

	// Handle UserDefined Case
	// Convert &v, iterate only once
	if t.Kind() != reflect.Ptr && allowAddr {
		if reflect.PtrTo(t).Implements(converterType) {
			return newCondAddrConvertFunc(addrUserDefinedConvertFunc, newTypeConverter(convFn, t, false))
		}
	}
	return convFn
}

// If CanAddr then get addr and handle else handle directly
type condAddrConvertFunc struct {
	canAddrConvert, elseConvert convertFunc
}

func (ce *condAddrConvertFunc) handle(v reflect.Value, tag reflect.StructTag) error {
	if v.CanAddr() {
		return ce.canAddrConvert(v, tag)
	}
	return ce.elseConvert(v, tag)
}

// newCondAddrConverter returns an encoder that checks whether its structTag
// CanAddr and delegates to canAddrConvert if so, else to elseConvert.
func newCondAddrConvertFunc(canAddrConvert, elseConvert convertFunc) convertFunc {
	convFn := &condAddrConvertFunc{canAddrConvert: canAddrConvert, elseConvert: elseConvert}
	return convFn.handle
}
