package default_

import (
	"reflect"
	"sync"
)

// map[reflect.Type]convertFunc

type converterCacheMap struct {
	converterCaches sync.Map
}

func (thiz *converterCacheMap) Store(reflectType reflect.Type, fn convertFunc) {
	thiz.converterCaches.Store(reflectType, fn)
}
func (thiz *converterCacheMap) LoadOrStore(reflectType reflect.Type, fn convertFunc) (convertFunc, bool) {
	actual, loaded := thiz.converterCaches.LoadOrStore(reflectType, fn)
	if actual == nil {
		return nil, loaded
	}
	return actual.(convertFunc), loaded
}
func (thiz *converterCacheMap) Load(reflectType reflect.Type) (convertFunc, bool) {
	fn, ok := thiz.converterCaches.Load(reflectType)
	if fn == nil {
		return nil, ok
	}
	return fn.(convertFunc), ok
}
func (thiz *converterCacheMap) Delete(reflectType reflect.Type) {
	thiz.converterCaches.Delete(reflectType)
}
func (thiz *converterCacheMap) Range(f func(reflectType reflect.Type, fn convertFunc) bool) {
	thiz.converterCaches.Range(func(reflectType, fn interface{}) bool {
		return f(reflectType.(reflect.Type), fn.(convertFunc))
	})
}
