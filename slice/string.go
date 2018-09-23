package slice

import "reflect"

func isAsRune(ifStringAsRune ...bool) bool {
	if ifStringAsRune != nil && len(ifStringAsRune) > 0 && ifStringAsRune[0] {
		return true
	}
	return false
}

func normalizeSlice(s []interface{}, as interface{}, ifStringAsRune ...bool) interface{} {
	if kind := reflect.ValueOf(as).Kind(); kind != reflect.String {
		return s
	}

	// AS []rune
	if isAsRune(ifStringAsRune...) {
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
