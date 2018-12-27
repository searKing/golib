package reflect_

import (
	"bytes"
	"fmt"
	"github.com/searKing/golib/bytes_"
	"reflect"
)

const PtrSize = 4 << (^uintptr(0) >> 63) // unsafe.Sizeof(uintptr(0)) but an ideal const, sizeof *void

func IsEmptyValue(v reflect.Value) bool {
	// return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
func IsZeroValue(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		if v.IsNil() {
			return true
		}
	default:
	}
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

func IsNilValue(v reflect.Value) (result bool) {
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	}
	return
}

func FollowValuePointer(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		return FollowValuePointer(v.Elem())
	}
	return v
}

// A field represents a single field found in a struct.
type fieldValueInfo struct {
	val   reflect.Value
	sf    reflect.StructField
	depth int
}

func (thiz *fieldValueInfo) String() string {
	//if IsNilValue(thiz.val) {
	//	return fmt.Sprintf("%+v", nil)
	//}
	//thiz.val.String()
	//return fmt.Sprintf("%+v %+v", thiz.val.Type().String(), thiz.val)

	switch k := thiz.val.Kind(); k {
	case reflect.Invalid:
		return "<invalid Value>"
	case reflect.String:
		return "[string: " + thiz.val.String() + "]"
	}
	// If you call String on a reflect.Value of other type, it's better to
	// print something than to panic. Useful in debugging.
	return "[" + thiz.val.Type().String() + ":" + fmt.Sprintf(" %+v", thiz.val) + "]"
}
func WalkValueDFS(val reflect.Value, parseFn func(info fieldValueInfo) (goon bool)) {
	// Types already visited at an earlier level.
	// FIXME I havenot seen any case which can trigger visited
	visited := map[reflect.Value]bool{}
	walkValueDFS(val, 0, visited, parseFn)
}

func walkValueDFS(val reflect.Value, depth int, visited map[reflect.Value]bool, parseFn func(info fieldValueInfo) (goon bool)) (goon bool) {
	if visited[val] {
		return true
	}
	visited[val] = true
	if !parseFn(fieldValueInfo{val: val, depth: depth}) {
		return false
	}
	if !val.IsValid() {
		return true
	}
	if IsNilType(val.Type()) {
		return true
	}

	if FollowValuePointer(val).Kind() != reflect.Struct {
		return true
	}
	// Scan typ for fields to include.
	for i := 0; i < val.NumField(); i++ {
		sf := val.Field(i)
		if !walkValueDFS(sf, depth+1, visited, parseFn) {
			return false
		}
	}
	return true
}

// Breadth First Search
func WalkValueBFS(val reflect.Value, parseFn func(info fieldValueInfo) (goon bool)) {
	// Types already visited at an earlier level.
	// FIXME I havenot seen any case which can trigger visited
	visited := map[reflect.Value]bool{}
	fvi := fieldValueInfo{val: val}
	if !parseFn(fvi) {
		return
	}
	walkValueBFS(fvi, visited, parseFn)
}
func walkValueBFS(fvi fieldValueInfo, visited map[reflect.Value]bool, parseFn func(info fieldValueInfo) (goon bool)) (goon bool) {
	depth := fvi.depth
	if visited[fvi.val] {
		return true
	}
	visited[fvi.val] = true
	if !fvi.val.IsValid() {
		return true
	}
	if IsNilType(fvi.val.Type()) {
		return true
	}
	if FollowValuePointer(fvi.val).Kind() != reflect.Struct {
		return true
	}
	// Anonymous fields to explore at the current level and the next.
	next := []reflect.Value{}

	depth++
	// Scan typ for fields to include.
	for i := 0; i < fvi.val.NumField(); i++ {
		sfv := fvi.val.Field(i)
		if visited[sfv] {
			continue
		}
		visited[sfv] = true

		if !parseFn(fieldValueInfo{val: sfv, sf: fvi.val.Type().Field(i), depth: depth}) {
			return false
		}

		next = append(next, sfv)
	}
	for _, sf := range next {
		if !walkValueDFS(sf, depth, visited, parseFn) {
			return false
		}
	}
	return true
}
func DumpValueInfoDFS(v reflect.Value) string {
	dumpInfo := &bytes.Buffer{}
	first := true
	WalkValueDFS(v, func(info fieldValueInfo) (goon bool) {
		if first {
			first = false
			bytes_.NewIndent(dumpInfo, "", "\t", info.depth)
		} else {
			bytes_.NewLine(dumpInfo, "", "\t", info.depth)
		}
		dumpInfo.WriteString(fmt.Sprintf("%+v", info.String()))
		return true
	})
	return dumpInfo.String()
}

func DumpValueInfoBFS(v reflect.Value) string {
	dumpInfo := &bytes.Buffer{}
	first := true
	WalkValueBFS(v, func(info fieldValueInfo) (goon bool) {
		if first {
			first = false
			bytes_.NewIndent(dumpInfo, "", "\t", info.depth)
		} else {
			bytes_.NewLine(dumpInfo, "", "\t", info.depth)
		}
		dumpInfo.WriteString(fmt.Sprintf("%+v", info.String()))
		return true
	})
	return dumpInfo.String()
}
