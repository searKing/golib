package reflect_

import (
	"reflect"
	"strings"
	"unicode"
)

// TagOptions is the string following a comma in a struct field's "json"
// tag, or the empty string. It does not include the leading comma.
type TagOptions string

// ParseTagOptions splits a struct field's json tag into its name and
// comma-separated options.
func ParseTagOptions(tag string) (tagName string, opts TagOptions) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], TagOptions(tag[idx+1:])
	}
	return tag, TagOptions("")
}

// Contains reports whether a comma-separated list of options
// contains a particular substr flag. substr must be surrounded by a
// string boundary or commas.
func (o TagOptions) Contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}
	s := string(o)
	for s != "" {
		var next string
		i := strings.Index(s, ",")
		if i >= 0 {
			s, next = s[:i], s[i+1:]
		}
		if s == optionName {
			return true
		}
		s = next
	}
	return false
}

func IsValidTagName(tagName string) bool {
	if tagName == "" {
		return false
	}
	for _, c := range tagName {
		switch {
		case strings.ContainsRune("!#$%&()*+-./:<=>?@[]^_{|}~ ", c):
			// Backslash and quote chars are reserved, but
			// otherwise any punctuation chars are allowed
			// in a tag name.
		default:
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
				return false
			}
		}
	}
	return true
}

// v[i][j][...], v[i] is v's ith field, v[i][j] is v[i]'s jth field
func FieldByStructIndex(v reflect.Value, index []int) reflect.Value {
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

// t[i][j][...], t[i] is t's ith field, t[i][j] is t[i]'s jth field
func TypeByStructFieldIndex(t reflect.Type, index []int) reflect.Type {
	for _, i := range index {
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		t = t.Field(i).Type
	}
	return t
}
