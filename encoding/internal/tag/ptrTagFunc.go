package tag

import "reflect"

type ptrTagFunc struct {
	elemConv tagFunc
}

func (pe *ptrTagFunc) handle(e *tagState, v reflect.Value, opts tagOpts) {
	if v.IsNil() {
		return
	}
	pe.elemConv(e, v.Elem(), opts)
}

func newPtrTagFunc(t reflect.Type) tagFunc {
	tagFn := &ptrTagFunc{typeConverter(t.Elem())}
	return tagFn.handle
}
