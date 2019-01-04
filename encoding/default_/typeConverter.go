package default_

import (
	"reflect"
	"sync"
)

func typeConverter(t reflect.Type) convertFunc {
	if fi, ok := converterCache.Load(t); ok {
		return fi
	}

	// To deal with recursive types, populate the map with an
	// indirect func before we build it. This type waits on the
	// real func (f) to be ready and then calls it. This indirect
	// func is only used for recursive types.
	var (
		wg sync.WaitGroup
		f  convertFunc
	)
	wg.Add(1)
	fi, loaded := converterCache.LoadOrStore(t, convertFunc(func(e *convertState, v reflect.Value, opts convOpts) {
		// wait until f is assigned elsewhere
		wg.Wait()
		f(e, v, opts)
	}))
	if loaded {
		return fi
	}

	// Compute the real encoder and replace the indirect func with it.
	f = newTypeConverter(t, true)
	wg.Done()
	converterCache.Store(t, f)
	return f
}

// newTypeEncoder constructs an convertorFunc for a type.
// The returned encoder only checks CanAddr when allowAddr is true.
func newTypeConverter(t reflect.Type, allowAddr bool) convertFunc {
	// Handle UserDefined Case
	// Convert v
	if t.Implements(converterType) {
		return userDefinedConverter
	}

	// Handle UserDefined Case
	// Convert &v, iterate only once
	if t.Kind() != reflect.Ptr && allowAddr {
		if reflect.PtrTo(t).Implements(converterType) {
			return newCondAddrConverter(addrUserDefinedConverter, newTypeConverter(t, false))
		}
	}

	// Handle BuiltinDefault Case
	switch t.Kind() {
	case reflect.Struct:
		return newStructConverter(t)
	case reflect.Ptr:
		return newPtrConverter(t)
	case reflect.Bool:
		fallthrough
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fallthrough
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		fallthrough
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		fallthrough
	case reflect.String:
		fallthrough
	case reflect.Interface:
		fallthrough
	case reflect.Map:
		fallthrough
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		fallthrough
	default:
		return newNopConverter(t)
		//return unsupportedTypeConverter
	}
}
