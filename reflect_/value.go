package reflect_

import (
	"bytes"
	"fmt"
	"github.com/searKing/golib/bytes_"
	"github.com/searKing/golib/container/traversal"
	"reflect"
)

const PtrSize = 4 << (^uintptr(0) >> 63) // unsafe.Sizeof(uintptr(0)) but an ideal const, sizeof *void

func IsEmptyValue(v reflect.Value) bool {
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
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
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
	v = reflect.Indirect(v)
	if v.Kind() == reflect.Ptr {
		return FollowValuePointer(v)
	}
	return v
}

// A field represents a single field found in a struct.
type FieldValueInfo struct {
	Value       reflect.Value
	StructField reflect.StructField
	Index       []int
}

func (thiz FieldValueInfo) Middles() []interface{} {

	if !thiz.Value.IsValid() {
		return nil
	}
	if IsNilType(thiz.Value.Type()) {
		return nil
	}
	val := FollowValuePointer(thiz.Value)
	if val.Kind() != reflect.Struct {
		return nil
	}

	middles := []interface{}{}
	// Scan typ for fields to include.
	for i := 0; i < val.NumField(); i++ {
		index := make([]int, len(thiz.Index)+1)
		copy(index, thiz.Index)
		index[len(thiz.Index)] = i
		middles = append(middles, FieldValueInfo{
			Value:       val.Field(i),
			StructField: val.Type().Field(i),
			Index:       index,
		})
	}
	return middles
}
func (thiz FieldValueInfo) Depth() int {
	return len(thiz.Index)
}

func (thiz *FieldValueInfo) String() string {
	//if IsNilValue(thiz.Value) {
	//	return fmt.Sprintf("%+v", nil)
	//}
	//thiz.Value.String()
	//return fmt.Sprintf("%+v %+v", thiz.Value.Type().String(), thiz.Value)

	switch k := thiz.Value.Kind(); k {
	case reflect.Invalid:
		return "<invalid Value>"
	case reflect.String:
		return "[string: " + thiz.Value.String() + "]"
	}
	// If you call String on a reflect.Value of other type, it's better to
	// print something than to panic. Useful in debugging.
	return "[" + thiz.Value.Type().String() + ":" + fmt.Sprintf(" %+v", thiz.Value) + "]"
}
func WalkValueDFS(val reflect.Value, parseFn func(info FieldValueInfo) (goon bool)) {
	traversal.TraversalBFS(FieldValueInfo{
		Value: val,
	}, nil, func(ele interface{}, depth int) (gotoNextLayer bool) {
		return parseFn(ele.(FieldValueInfo))
	})
}

// Breadth First Search
func WalkValueBFS(val reflect.Value, parseFn func(info FieldValueInfo) (goon bool)) {
	traversal.TraversalBFS(FieldValueInfo{Value: val},
		nil, func(ele interface{}, depth int) (gotoNextLayer bool) {
			return parseFn(ele.(FieldValueInfo))
		})
}

func DumpValueInfoDFS(v reflect.Value) string {
	dumpInfo := &bytes.Buffer{}
	first := true
	WalkValueDFS(v, func(info FieldValueInfo) (goon bool) {
		if first {
			first = false
			bytes_.NewIndent(dumpInfo, "", "\t", info.Depth())
		} else {
			bytes_.NewLine(dumpInfo, "", "\t", info.Depth())
		}
		dumpInfo.WriteString(fmt.Sprintf("%+v", info.String()))
		return true
	})
	return dumpInfo.String()
}

func DumpValueInfoBFS(v reflect.Value) string {
	dumpInfo := &bytes.Buffer{}
	first := true
	WalkValueBFS(v, func(info FieldValueInfo) (goon bool) {
		if first {
			first = false
			bytes_.NewIndent(dumpInfo, "", "\t", info.Depth())
		} else {
			bytes_.NewLine(dumpInfo, "", "\t", info.Depth())
		}
		dumpInfo.WriteString(fmt.Sprintf("%+v", info.String()))
		return true
	})
	return dumpInfo.String()
}
