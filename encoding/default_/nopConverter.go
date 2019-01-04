package default_

import "reflect"

type nopConverter struct {
}

func (_ *nopConverter) convert(e *convertState, v reflect.Value, opts convOpts) {
	// nop
	return
}

func newNopConverter(t reflect.Type) convertFunc {
	conv := &nopConverter{}
	return conv.convert
}
