package reflect_

import (
	"bytes"
	"fmt"
	"github.com/searKing/golib/bytes_"
	"github.com/searKing/golib/container/traversal"
	"reflect"
)

// nil, unknown type
func IsNilType(v reflect.Type) (result bool) {
	return v == nil
}
func FollowTypePointer(v reflect.Type) reflect.Type {
	if IsNilType(v) {
		return v
	}
	if v.Kind() == reflect.Ptr {
		return FollowTypePointer(v.Elem())
	}
	return v
}

// A field represents a single field found in a struct.
type FieldTypeInfo struct {
	StructField reflect.StructField
	Depth       int
	Index       []int
}

func (thiz FieldTypeInfo) Middles() []interface{} {
	typ := thiz.StructField.Type
	middles := []interface{}{}
	typ = FollowTypePointer(typ)
	if IsNilType(typ) {
		return nil
	}
	if typ.Kind() != reflect.Struct {
		return nil
	}
	// Scan typ for fields to include.
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		middles = append(middles, FieldTypeInfo{
			StructField: sf,
			Depth:       thiz.Depth + 1,
			Index:       append(thiz.Index, i),
		})
	}
	return middles
}

func (thiz FieldTypeInfo) String() string {
	if thiz.StructField.Type == nil {
		return fmt.Sprintf("%+v", nil)
	}
	return fmt.Sprintf("%+v", thiz.StructField.Type.String())
}

// Breadth First Search
func WalkTypeBFS(typ reflect.Type, parseFn func(info FieldTypeInfo) (goon bool)) {
	traversal.TraversalBFS(FieldTypeInfo{
		StructField: reflect.StructField{
			Type: typ,
		},
	}, nil, func(ele interface{}, depth int) (gotoNextLayer bool) {
		return parseFn(ele.(FieldTypeInfo))
	})
}

// Wid First Search
func WalkTypeDFS(typ reflect.Type, parseFn func(info FieldTypeInfo) (goon bool)) {
	traversal.TraversalDFS(FieldTypeInfo{
		StructField: reflect.StructField{
			Type: typ,
		},
	}, nil, func(ele interface{}, depth int) (gotoNextLayer bool) {
		return parseFn(ele.(FieldTypeInfo))
	})
}
func DumpTypeInfoDFS(t reflect.Type) string {
	dumpInfo := &bytes.Buffer{}
	first := true
	WalkTypeDFS(t, func(info FieldTypeInfo) (goon bool) {
		if first {
			first = false
			bytes_.NewIndent(dumpInfo, "", "\t", info.Depth)
		} else {
			bytes_.NewLine(dumpInfo, "", "\t", info.Depth)
		}
		dumpInfo.WriteString(fmt.Sprintf("%+v", info.String()))
		return true
	})
	return dumpInfo.String()
}
func DumpTypeInfoBFS(t reflect.Type) string {
	dumpInfo := &bytes.Buffer{}
	first := true
	WalkTypeBFS(t, func(info FieldTypeInfo) (goon bool) {
		if first {
			first = false
			bytes_.NewIndent(dumpInfo, "", "\t", info.Depth)
		} else {
			bytes_.NewLine(dumpInfo, "", "\t", info.Depth)
		}
		dumpInfo.WriteString(fmt.Sprintf("%+v", info.String()))
		return true
	})
	return dumpInfo.String()
}
