package tag

import "reflect"

// If CanAddr then get addr and handle else handle directly
type condAddrTagFunc struct {
	canAddrTagFunc, elseTagFunc tagFunc
}

func (ce *condAddrTagFunc) handle(e *tagState, v reflect.Value, opts tagOpts) {
	if v.CanAddr() {
		ce.canAddrTagFunc(e, v, opts)
	} else {
		ce.elseTagFunc(e, v, opts)
	}
}

// newCondAddrConverter returns an encoder that checks whether its structTag
// CanAddr and delegates to canAddrTagFunc if so, else to elseTagFunc.
func newCondAddrTagFunc(canAddrConvert, elseConvert tagFunc) tagFunc {
	tagFn := &condAddrTagFunc{canAddrTagFunc: canAddrConvert, elseTagFunc: elseConvert}
	return tagFn.handle
}
