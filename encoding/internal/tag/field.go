package tag

import (
	"github.com/searKing/golib/reflect_"
	"reflect"
	"sync"
)

// A field represents a single field found in a struct.
type field struct {
	name      string
	structTag reflect.StructTag

	index []int
	typ   reflect.Type
}

var fieldCache fieldMap

type fieldMap struct {
	fields sync.Map // map[reflect.Type][]field
}

func (thiz *fieldMap) Store(type_ reflect.Type, fields []field) {
	thiz.fields.Store(type_, fields)
}

func (thiz *fieldMap) LoadOrStore(type_ reflect.Type, fields []field) ([]field, bool) {
	actual, loaded := thiz.fields.LoadOrStore(type_, fields)
	if actual == nil {
		return nil, loaded
	}
	return actual.([]field), loaded
}

func (thiz *fieldMap) Load(type_ reflect.Type) ([]field, bool) {
	fields, ok := thiz.fields.Load(type_)
	if fields == nil {
		return nil, ok
	}
	return fields.([]field), ok
}

func (thiz *fieldMap) Delete(type_ reflect.Type) {
	thiz.fields.Delete(type_)
}

func (thiz *fieldMap) Range(f func(type_ reflect.Type, fields []field) bool) {
	thiz.fields.Range(func(type_, fields interface{}) bool {
		return f(type_.(reflect.Type), fields.([]field))
	})
}

// cachedTypeFields is like typeFields but uses a cache to avoid repeated work.
func cachedTypeFields(t reflect.Type) []field {
	if f, ok := fieldCache.Load(t); ok {
		return f
	}
	fields := []field{}
	reflect_.WalkTypeDFS(t, func(info reflect_.FieldTypeInfo) (goon bool) {
		// ignore struct's root
		if info.Depth() == 0 {
			return true
		}

		sf, ok := info.StructField()
		if !ok {
			return true
		}

		fields = append(fields, field{
			name:      sf.Name,
			structTag: sf.Tag,
			index:     info.Index(),
			typ:       sf.Type,
		})
		return true
	})
	f, _ := fieldCache.LoadOrStore(t, fields)
	return f
}
