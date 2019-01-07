package tag

import "reflect"

// If CanAddr then get addr and handle else handle directly
type condAddrTagFunc struct {
	canAddrConvert, elseConvert tagFunc
}

func (ce *condAddrTagFunc) handle(e *tagState, v reflect.Value, opts tagOpts) {
	if v.CanAddr() {
		ce.canAddrConvert(e, v, opts)
	} else {
		ce.elseConvert(e, v, opts)
	}
}

// newCondAddrConverter returns an encoder that checks whether its structTag
// CanAddr and delegates to canAddrConvert if so, else to elseConvert.
func newCondAddrTagFunc(canAddrConvert, elseConvert tagFunc) tagFunc {
	tagFn := &condAddrTagFunc{canAddrConvert: canAddrConvert, elseConvert: elseConvert}
	return tagFn.handle
}
