package reflect_

import (
	"bytes"
	"fmt"
	"github.com/searKing/golib/bytes_"
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
	sf     reflect.StructField
	deepth int
}

func (thiz *fieldTypeInfo) String() string {
	if thiz.sf.Type == nil {
		return fmt.Sprintf("%+v", nil)
	}
	return fmt.Sprintf("%+v", thiz.sf.Type.String())
}

// Depth First Search
func WalkTypeDFS(typ reflect.Type, parseFn func(info fieldTypeInfo) (goon bool)) {
	// Types already visited at an earlier level.
	// FIXME I havenot seen any case which can trigger visited
	visited := map[reflect.Type]bool{}
	walkTypeDFS(reflect.StructField{
		Type: typ,
	}, 0, visited, parseFn)
}
func walkTypeDFS(sf reflect.StructField, deepth int, visited map[reflect.Type]bool, parseFn func(info fieldTypeInfo) (goon bool)) (goon bool) {
	typ := sf.Type
	if visited[typ] {
		return true
	}
	visited[typ] = true
	if !parseFn(fieldTypeInfo{sf: sf, deepth: deepth}) {
		return false
	}
	typ = FollowTypePointer(typ)
	if IsNilType(typ) {
		return true
	}
	if typ.Kind() != reflect.Struct {
		return true
	}
	// Scan typ for fields to include.
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		//ft := sf.Type
		//// *BaseStruct
		//if ft.Name() == "" && ft.Kind() == reflect.Ptr {
		//	// Follow pointer.
		//	ft = FollowTypePointer(ft)
		//}
		if !walkTypeDFS(sf, deepth+1, visited, parseFn) {
			return false
		}
	}
	return true
}

// Breadth First Search
func WalkTypeBFS(typ reflect.Type, parseFn func(info fieldTypeInfo) (goon bool)) {
	// Types already visited at an earlier level.
	// FIXME I havenot seen any case which can trigger visited
	visited := map[reflect.Type]bool{}
	fti := fieldTypeInfo{
		sf: reflect.StructField{
			Type: typ,
		},
		deepth: 0,
	}
	if !parseFn(fti) {
		return
	}
	walkTypeBFS(fti, visited, parseFn)
}
func walkTypeBFS(fti fieldTypeInfo, visited map[reflect.Type]bool, parseFn func(info fieldTypeInfo) (goon bool)) (goon bool) {
	typ := fti.sf.Type
	if visited[typ] {
		return true
	}
	visited[typ] = true
	typ = FollowTypePointer(typ)
	if IsNilType(typ) {
		return true
	}
	if typ.Kind() != reflect.Struct {
		return true
	}
	// Anonymous fields to explore at the current level and the next.
	next := []reflect.StructField{}

	fti.deepth++
	// Scan typ for fields to include.
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		ft := sf.Type
		//// *BaseStruct
		//if ft.Name() == "" && ft.Kind() == reflect.Ptr {
		//	// Follow pointer.
		//	ft = FollowTypePointer(ft)
		//}
		if visited[ft] {
			continue
		}
		visited[ft] = true

		if !parseFn(fieldTypeInfo{sf: sf, deepth: fti.deepth}) {
			return false
		}

		next = append(next, sf)
	}
	for _, sf := range next {
		if !walkTypeBFS(fieldTypeInfo{sf: sf, deepth: fti.deepth}, visited, parseFn) {
			return false
		}
	}
	return true
}
func DumpTypeInfoDFS(t reflect.Type) string {
	dumpInfo := &bytes.Buffer{}
	first := true
	WalkTypeDFS(t, func(info fieldTypeInfo) (goon bool) {
		if first {
			first = false
			bytes_.NewIndent(dumpInfo, "", "\t", info.deepth)
		} else {
			bytes_.NewLine(dumpInfo, "", "\t", info.deepth)
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
			bytes_.NewIndent(dumpInfo, "", "\t", info.deepth)
		} else {
			bytes_.NewLine(dumpInfo, "", "\t", info.deepth)
		}
		dumpInfo.WriteString(fmt.Sprintf("%+v", info.String()))
		return true
	})
	return dumpInfo.String()
}
