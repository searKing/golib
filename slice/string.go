package slice

import "reflect"

func isAsRune(ifStringAsRune ...bool) bool {
	if ifStringAsRune != nil && len(ifStringAsRune) > 0 && ifStringAsRune[0] {
		return true
	}
	return false
}

func normalizeSlice(s []interface{}, as interface{}, ifStringAsRune ...bool) interface{} {
	kind := reflect.ValueOf(as).Kind()
	switch kind {
	case reflect.Map:
		return normalizeSliceAsMap(s)
	case reflect.String:
		return normalizeSliceAsString(s, isAsRune(ifStringAsRune...))

	}
	return s
}
func normalizeElem(elem, as interface{}, ifStringAsRune ...bool) interface{} {
	if kind := reflect.ValueOf(as).Kind(); kind != reflect.String {
		return elem
	}

	// AS rune
	if isAsRune(ifStringAsRune...) {
		return string([]rune{elem.(rune)})
	}
	// AS byte
	return string([]byte{elem.(byte)})
}

func normalizeSliceAsString(s []interface{}, asRune bool) interface{} {
	// AS []rune
	if asRune {
		bs := make([]rune, 0, len(s))
		for _, s := range s {
			bs = append(bs, s.(rune))
		}
		return string(bs)
	}
	// AS []byte
	bs := make([]byte, len(s))
	for _, s := range s {
		bs = append(bs, s.(byte))
	}
	return string(bs)
}

func normalizeSliceAsMap(s []interface{}) interface{} {
	bs := make(map[interface{}]interface{})
	for _, m := range s {
		pair := m.(MapPair)
		bs[pair.key] = pair.value
	}
	return bs
}
