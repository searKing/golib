package tag

import (
	"github.com/searKing/golib/reflect_"
	"reflect"
)

type structTagFunc struct {
	fields     []field
	fieldConvs []tagFunc
}

func (se *structTagFunc) handle(state *tagState, v reflect.Value, opts tagOpts) {
	for i, f := range se.fields {
		fv := reflect_.ValueByStructFieldIndex(v, f.index)
		if !fv.IsValid() && reflect_.IsEmptyValue(fv) {
			continue
		}
		field := v.FieldByIndex(se.fields[i].index)
		//if field.Type().Implements(taggerType) {
		//	field.Interface().(Tagger).TagDefault()
		//	continue
		//}

		//判断是否为可取指，可导出字段
		if !field.CanAddr() || !field.CanInterface() {
			continue
		}

		tagFn := opts.TagHandler
		if tagFn != nil {
			if err := tagFn(field, se.fields[i].structTag); err != nil {
				state.error(&TaggerError{v.Type(), err})
				return
			}
		}
		se.fieldConvs[i](state, fv, opts)
	}
}

func newStructTagFunc(t reflect.Type) tagFunc {
	fields := cachedTypeFields(t)
	se := &structTagFunc{
		fields:     fields,
		fieldConvs: make([]tagFunc, len(fields)),
	}
	for i, f := range fields {
		se.fieldConvs[i] = typeConverter(reflect_.TypeByStructFieldIndex(t, f.index))
	}
	return se.handle
}
