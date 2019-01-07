package default_

import (
	"github.com/searKing/golib/reflect_"
	"gopkg.in/yaml.v2"
	"reflect"
)

// A field represents a single field found in a struct.
type field struct {
	name  string
	value string

	index []int

	tag bool
	typ reflect.Type
}
type structConverter struct {
	fields     []field
	fieldConvs []convertFunc
}

func (se *structConverter) convert(e *convertState, v reflect.Value, opts convOpts) {
	for i, f := range se.fields {
		fv := reflect_.FieldByStructIndex(v, f.index)
		if !fv.IsValid() && reflect_.IsEmptyValue(fv) {
			continue
		}
		if se.fields[i].tag {
			field := v.FieldByIndex(se.fields[i].index)
			//判断是否为可取指，可导出字段
			if !field.CanAddr() || !field.CanInterface() {
				continue
			}
			if err := yaml.Unmarshal([]byte(se.fields[i].value), field.Addr().Interface()); err != nil {
				e.error(&ConverterError{v.Type(), err})
				return
			}
		}
		se.fieldConvs[i](e, fv, opts)
	}
}

func newStructConverter(t reflect.Type) convertFunc {
	fields := cachedTypeFields(t)
	se := &structConverter{
		fields:     fields,
		fieldConvs: make([]convertFunc, len(fields)),
	}
	for i, f := range fields {
		se.fieldConvs[i] = typeConverter(reflect_.TypeByStructFieldIndex(t, f.index))
	}
	return se.convert
}
