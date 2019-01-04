package default_

import "reflect"

type ptrConverter struct {
	elemConv convertFunc
}

func (pe *ptrConverter) convert(e *convertState, v reflect.Value, opts convOpts) {
	if v.IsNil() {
		return
	}
	pe.elemConv(e, v.Elem(), opts)
}

func newPtrConverter(t reflect.Type) convertFunc {
	conv := &ptrConverter{typeConverter(t.Elem())}
	return conv.convert
}
