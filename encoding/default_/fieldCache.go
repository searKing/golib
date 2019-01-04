package default_

import (
	"github.com/searKing/golib/bufio_"
	"github.com/searKing/golib/reflect_"
	"reflect"
	"strings"
	"sync"
)

var fieldCache FieldCacheMap

type FieldCacheMap struct {
	fieldCaches sync.Map // map[reflect.Type][]field
}

func (thiz *FieldCacheMap) Store(type_ reflect.Type, fields []field) {
	thiz.fieldCaches.Store(type_, fields)
}
func (thiz *FieldCacheMap) LoadOrStore(type_ reflect.Type, fields []field) ([]field, bool) {
	actual, loaded := thiz.fieldCaches.LoadOrStore(type_, fields)
	if actual == nil {
		return nil, loaded
	}
	return actual.([]field), loaded
}
func (thiz *FieldCacheMap) Load(type_ reflect.Type) ([]field, bool) {
	fields, ok := thiz.fieldCaches.Load(type_)
	if fields == nil {
		return nil, ok
	}
	return fields.([]field), ok
}
func (thiz *FieldCacheMap) Delete(type_ reflect.Type) {
	thiz.fieldCaches.Delete(type_)
}
func (thiz *FieldCacheMap) Range(f func(type_ reflect.Type, fields []field) bool) {
	thiz.fieldCaches.Range(func(type_, fields interface{}) bool {
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
		sf := info.StructField
		if reflect_.IsEmptyValue(reflect.ValueOf(sf)) {
			return true
		}
		// Handle Tag
		tag := sf.Tag.Get("default")
		if tag == "" {
			return true
		}
		if strings.TrimSpace(tag) == "-" {
			return true
		}
		kind := sf.Type.Kind()
		if kind == reflect.Slice ||
			kind == reflect.Map ||
			kind == reflect.Array {
			if tag[0] == '[' || tag[0] == '{' {
				defaultValue, err := bufio_.NewPairScanner(strings.NewReader(tag)).SetDiscardLeading(true).ScanDelimiters("{}[]")
				if err == nil {
					tag = string(defaultValue)
				}
			}
		}

		fields = append(fields, field{
			name:  sf.Name,
			value: tag,
			index: info.Index,
			tag:   true,
			typ:   sf.Type,
		})
		return true
	})
	f, _ := fieldCache.LoadOrStore(t, fields)
	return f
}
