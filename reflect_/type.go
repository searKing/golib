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
type fieldTypeInfo struct {
	sf    reflect.StructField
	depth int
}

func (thiz fieldTypeInfo) Middles() []interface{} {
	typ := thiz.sf.Type
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
		middles = append(middles, fieldTypeInfo{
			sf:    sf,
			depth: thiz.depth + 1,
		})
	}
	return middles
}

func (thiz fieldTypeInfo) String() string {
	if thiz.sf.Type == nil {
		return fmt.Sprintf("%+v", nil)
	}
	return fmt.Sprintf("%+v", thiz.sf.Type.String())
}

// Breadth First Search
func WalkTypeBFS(typ reflect.Type, parseFn func(info fieldTypeInfo) (goon bool)) {
	traversal.TraversalBFS(fieldTypeInfo{
		sf: reflect.StructField{
			Type: typ,
		},
		depth: 0,
	}, nil, func(ele interface{}, depth int) (gotoNextLayer bool) {
		return parseFn(ele.(fieldTypeInfo))
	})
}

// Wid First Search
func WalkTypeDFS(typ reflect.Type, parseFn func(info fieldTypeInfo) (goon bool)) {
	traversal.TraversalDFS(fieldTypeInfo{
		sf: reflect.StructField{
			Type: typ,
		},
		depth: 0,
	}, nil, func(ele interface{}, depth int) (gotoNextLayer bool) {
		return parseFn(ele.(fieldTypeInfo))
	})
}
func DumpTypeInfoDFS(t reflect.Type) string {
	dumpInfo := &bytes.Buffer{}
	first := true
	WalkTypeDFS(t, func(info fieldTypeInfo) (goon bool) {
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
func DumpTypeInfoBFS(t reflect.Type) string {
	dumpInfo := &bytes.Buffer{}
	first := true
	WalkTypeBFS(t, func(info fieldTypeInfo) (goon bool) {
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
