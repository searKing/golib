package tag

import "reflect"

type nopTagFunc struct {
}

func (_ *nopTagFunc) handle(e *tagState, v reflect.Value, opts tagOpts) {
	// nop
	return
}

func newNopConverter(t reflect.Type) tagFunc {
	tagFn := &nopTagFunc{}
	return tagFn.handle
}
