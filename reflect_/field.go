package reflect_

import (
	"reflect"
)

// A field represents a single field found in a struct.
type field struct {
	sf reflect.StructField

	name      string
	nameBytes []byte                 // []byte(name)
	equalFold func(s, t []byte) bool // bytes.EqualFold or equivalent

	tag       bool
	index     []int
	typ       reflect.Type
	omitEmpty bool
	quoted    bool
}

//
//// typeFields returns a list of fields that JSON should recognize for the given type.
//// The algorithm is breadth-first search over the set of structs to include - the top struct
//// and then any reachable anonymous structs.
//func typeFields(t reflect.Type, parseTag func(tag reflect.StructTag) (stop bool)) []field {
//	// Anonymous fields to explore at the current level and the next.
//	current := []field{}
//	next := []field{{typ: t}}
//
//	// Count of queued names for current level and the next.
//	currentCount := map[reflect.Type]int{}
//	nextCount := map[reflect.Type]int{}
//
//	// Types already visited at an earlier level.
//	visited := map[reflect.Type]bool{}
//
//	// Fields found.
//	var fields []field
//
//	for len(next) > 0 {
//		current, next = next, current[:0]
//		currentCount, nextCount = nextCount, map[reflect.Type]int{}
//
//		for _, f := range current {
//			// make sure field is visited only once
//			if visited[f.typ] {
//				continue
//			}
//			visited[f.typ] = true
//
//			// Scan f.typ for fields to include.
//			for i := 0; i < f.typ.NumField(); i++ {
//				StructField := f.typ.Field(i)
//				isUnexported := StructField.PkgPath != ""
//				if StructField.Anonymous {
//					t := StructField.Type
//					if t.Kind() == reflect.Ptr {
//						t = t.Elem()
//					}
//					if isUnexported && t.Kind() != reflect.Struct {
//						// Ignore embedded fields of unexported non-struct types.
//						continue
//					}
//					// Do not ignore embedded fields of unexported struct types
//					// since they may have exported fields.
//				} else if isUnexported {
//					// Ignore unexported non-embedded fields.
//					continue
//				}
//				if parseTag(StructField.Tag) {
//					continue
//				}
//				index := make([]int, len(f.index)+1)
//				copy(index, f.index)
//				index[len(f.index)] = i
//
//				ft := StructField.Type
//				if ft.Name() == "" && ft.Kind() == reflect.Ptr {
//					// Follow pointer.
//					ft = ft.Elem()
//				}
//
//				// Record found field and index sequence.
//				if name != "" || !StructField.Anonymous || ft.Kind() != reflect.Struct {
//					tagged := name != ""
//					if name == "" {
//						name = StructField.Name
//					}
//					fields = append(fields, fillField(field{
//						name:      name,
//						tag:       tagged,
//						index:     index,
//						typ:       ft,
//						omitEmpty: opts.Contains("omitempty"),
//						quoted:    quoted,
//					}))
//					if currentCount[f.typ] > 1 {
//						// If there were multiple instances, add a second,
//						// so that the annihilation code will see a duplicate.
//						// It only cares about the distinction between 1 or 2,
//						// so don't bother generating any more copies.
//						fields = append(fields, fields[len(fields)-1])
//					}
//					continue
//				}
//
//				// Record new anonymous struct to explore in next round.
//				nextCount[ft]++
//				if nextCount[ft] == 1 {
//					next = append(next, fillField(field{name: ft.Name(), index: index, typ: ft}))
//				}
//			}
//		}
//	}
//
//	sort.Slice(fields, func(i, j int) bool {
//		x := fields
//		// sort field by name, breaking ties with Depth, then
//		// breaking ties with "name came from json tag", then
//		// breaking ties with index sequence.
//		if x[i].name != x[j].name {
//			return x[i].name < x[j].name
//		}
//		if len(x[i].index) != len(x[j].index) {
//			return len(x[i].index) < len(x[j].index)
//		}
//		if x[i].tag != x[j].tag {
//			return x[i].tag
//		}
//		return byIndex(x).Less(i, j)
//	})
//
//	// Delete all fields that are hidden by the Go rules for embedded fields,
//	// except that fields with JSON tags are promoted.
//
//	// The fields are sorted in primary order of name, secondary order
//	// of field index length. Loop over names; for each name, delete
//	// hidden fields by choosing the one dominant field that survives.
//	out := fields[:0]
//	for advance, i := 0, 0; i < len(fields); i += advance {
//		// One iteration per name.
//		// Find the sequence of fields with the name of this first field.
//		fi := fields[i]
//		name := fi.name
//		for advance = 1; i+advance < len(fields); advance++ {
//			fj := fields[i+advance]
//			if fj.name != name {
//				break
//			}
//		}
//		if advance == 1 { // Only one field with this name
//			out = append(out, fi)
//			continue
//		}
//		dominant, ok := dominantField(fields[i : i+advance])
//		if ok {
//			out = append(out, dominant)
//		}
//	}
//
//	fields = out
//	sort.Sort(byIndex(fields))
//
//	return fields
//}
//
//func fillField(f field) field {
//	f.nameBytes = []byte(f.name)
//	f.equalFold = foldFunc(f.nameBytes)
//	return f
//}
//func parseStructField(StructField reflect.StructField) (stop bool) {
//
//	isUnexported := StructField.PkgPath != ""
//	if StructField.Anonymous {
//		t := StructField.Type
//		if t.Kind() == reflect.Ptr {
//			t = t.Elem()
//		}
//		if isUnexported && t.Kind() != reflect.Struct {
//			// Ignore embedded fields of unexported non-struct types.
//			return false
//		}
//		// Do not ignore embedded fields of unexported struct types
//		// since they may have exported fields.
//	} else if isUnexported {
//		// Ignore unexported non-embedded fields.
//		return false
//	}
//	// Record found field and index sequence.
//	if name != "" || !StructField.Anonymous || ft.Kind() != reflect.Struct {
//		tagged := name != ""
//		if name == "" {
//			name = StructField.Name
//		}
//		fields = append(fields, fillField(field{
//			name:      name,
//			tag:       tagged,
//			index:     index,
//			typ:       ft,
//			omitEmpty: opts.Contains("omitempty"),
//			quoted:    quoted,
//		}))
//		if currentCount[f.typ] > 1 {
//			// If there were multiple instances, add a second,
//			// so that the annihilation code will see a duplicate.
//			// It only cares about the distinction between 1 or 2,
//			// so don't bother generating any more copies.
//			fields = append(fields, fields[len(fields)-1])
//		}
//		continue
//	}
//
//}

//func workFields(t reflect.Type, do func(StructField reflect.StructField) (stop bool)) {
//	// Anonymous fields to explore at the current level and the next.
//	current := []field{}
//	next := []field{{typ: t}}
//
//	// Count of queued names for current level and the next.
//	currentCount := map[reflect.Type]int{}
//	nextCount := map[reflect.Type]int{}
//
//	// Types already visited at an earlier level.
//	visited := map[reflect.Type]bool{}
//
//	// Fields found.
//	var fields []field
//
//	for len(next) > 0 {
//		current, next = next, current[:0]
//		currentCount, nextCount = nextCount, map[reflect.Type]int{}
//
//		for _, f := range current {
//			// make sure field is visited only once
//			if visited[f.typ] {
//				continue
//			}
//			visited[f.typ] = true
//
//			if f.typ.Kind() == reflect.Struct || f.typ.Elem().Kind() == reflect.Struct {
//
//			}
//			// Scan f.typ for fields to include.
//			for i := 0; i < f.typ.NumField(); i++ {
//				StructField := f.typ.Field(i)
//				if do(StructField) {
//					continue
//				}
//				index := make([]int, len(f.index)+1)
//				copy(index, f.index)
//				index[len(f.index)] = i
//
//				ft := StructField.Type
//				if ft.Name() == "" && ft.Kind() == reflect.Ptr {
//					// Follow pointer.
//					ft = ft.Elem()
//				}
//
//				// Record new anonymous struct to explore in next round.
//				nextCount[ft]++
//				if nextCount[ft] == 1 {
//					next = append(next, fillField(field{name: ft.Name(), index: index, typ: ft}))
//				}
//			}
//		}
//	}
//
//}
// v[i,j,k...] of struct
func FieldByIndex(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return reflect.Value{}
			}
			v = v.Elem()
		}
		v = v.Field(i)
	}
	return v
}

func IsFieldExported(sf reflect.StructField) bool {
	return sf.PkgPath == ""
}
