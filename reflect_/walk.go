package reflect_

import (
	"reflect"
)

// walks down v
func Walk(t reflect.Type, visitedOnce bool, do func(s reflect.Type, sf reflect.StructField) (stop bool)) {
	// Anonymous fields to explore at the current level and the next.
	current := []reflect.Type{}
	next := []reflect.Type{t}

	// Count of queued names for current level and the next.
	currentCount := map[reflect.Type]int{}
	nextCount := map[reflect.Type]int{}

	// Types already visited at an earlier level.
	// FIXME I havenot seen any case which can trigger visited
	visited := map[reflect.Type]bool{}
	for len(next) > 0 {
		current, next = next, current[:0]
		currentCount, nextCount = nextCount, map[reflect.Type]int{}

		for _, typ := range current {

			if typ.Kind() == reflect.Ptr {
				// Follow pointer.
				typ = typ.Elem()
			}
			if visitedOnce {
				if visited[typ] {
					continue
				}
				visited[typ] = true
			}

			if typ.Kind() != reflect.Struct {
				if do(typ, reflect.StructField{}) {
					return
				}
				continue
			}
			// Scan typ for fields to include.
			for i := 0; i < typ.NumField(); i++ {
				sf := typ.Field(i)
				if do(typ, sf) {
					continue
				}

				ft := sf.Type
				if ft.Name() == "" && ft.Kind() == reflect.Ptr {
					// Follow pointer.
					ft = ft.Elem()
				}

				// Record found field and index sequence.
				if ft.Name() != "" || !sf.Anonymous || ft.Kind() != reflect.Struct {
					if currentCount[typ] > 1 {
					}
					//continue
				}
				// Record new anonymous struct to explore in next round.
				nextCount[ft]++
				if !visitedOnce || nextCount[ft] == 1 {
					next = append(next, ft)
				}
			}
		}
	}

}

//
//// indirect walks down v allocating pointers as needed,
//// until it gets to a non-pointer.
//// if it encounters an Unmarshaler, indirect stops and returns that.
//// if decodingNull is true, indirect stops at the last pointer so it can be set to nil.
//func indirect(v reflect.Value, decodingNull bool) (Unmarshaler, encoding.TextUnmarshaler, reflect.Value) {
//	// Issue #24153 indicates that it is generally not a guaranteed property
//	// that you may round-trip a reflect.Value by calling Value.Addr().Elem()
//	// and expect the value to still be settable for values derived from
//	// unexported embedded struct fields.
//	//
//	// The logic below effectively does this when it first addresses the value
//	// (to satisfy possible pointer methods) and continues to dereference
//	// subsequent pointers as necessary.
//	//
//	// After the first round-trip, we set v back to the original value to
//	// preserve the original RW flags contained in reflect.Value.
//	v0 := v
//	haveAddr := false
//
//	// If v is a named type and is addressable,
//	// start with its address, so that if the type has pointer methods,
//	// we find them.
//	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
//		haveAddr = true
//		v = v.Addr()
//	}
//	for {
//		// Load value from interface, but only if the result will be
//		// usefully addressable.
//		if v.Kind() == reflect.Interface && !v.IsNil() {
//			e := v.Elem()
//			if e.Kind() == reflect.Ptr && !e.IsNil() && (!decodingNull || e.Elem().Kind() == reflect.Ptr) {
//				haveAddr = false
//				v = e
//				continue
//			}
//		}
//
//		if v.Kind() != reflect.Ptr {
//			break
//		}
//
//		if v.Elem().Kind() != reflect.Ptr && decodingNull && v.CanSet() {
//			break
//		}
//		if v.IsNil() {
//			v.Set(reflect.New(v.Type().Elem()))
//		}
//		if v.Type().NumMethod() > 0 {
//			if u, ok := v.Interface().(Unmarshaler); ok {
//				return u, nil, reflect.Value{}
//			}
//			if !decodingNull {
//				if u, ok := v.Interface().(encoding.TextUnmarshaler); ok {
//					return nil, u, reflect.Value{}
//				}
//			}
//		}
//
//		if haveAddr {
//			v = v0 // restore original value after round-trip Value.Addr().Elem()
//			haveAddr = false
//		} else {
//			v = v.Elem()
//		}
//	}
//	return nil, nil, v
//}
