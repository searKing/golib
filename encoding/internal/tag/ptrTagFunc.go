package tag

import "reflect"

type ptrTagFunc struct {
	elemConv tagFunc
}

func (pe *ptrTagFunc) handle(e *tagState, v reflect.Value, opts tagOpts) (isUserDefined bool) {
	isUserDefined = false
	if v.IsNil() {
		return
	}
	return pe.elemConv(e, v.Elem(), opts)
}

func newPtrTagFunc(t reflect.Type) tagFunc {
	tagFn := &ptrTagFunc{typeTagFunc(t.Elem())}
	return tagFn.handle
}
