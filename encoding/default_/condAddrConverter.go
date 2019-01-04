package default_

import "reflect"

// If CanAddr then get addr and convert else convert directly
type condAddrConverter struct {
	canAddrConvert, elseConvert convertFunc
}

func (ce *condAddrConverter) convert(e *convertState, v reflect.Value, opts convOpts) {
	if v.CanAddr() {
		ce.canAddrConvert(e, v, opts)
	} else {
		ce.elseConvert(e, v, opts)
	}
}

// newCondAddrConverter returns an encoder that checks whether its value
// CanAddr and delegates to canAddrConvert if so, else to elseConvert.
func newCondAddrConverter(canAddrConvert, elseConvert convertFunc) convertFunc {
	conv := &condAddrConverter{canAddrConvert: canAddrConvert, elseConvert: elseConvert}
	return conv.convert
}
