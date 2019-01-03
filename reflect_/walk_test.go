package reflect_

import (
	"reflect"
	"testing"
)

type output struct {
	a      *bool `json:"MemberA"`
	b      bool
	c      []bool
	expect bool
}
type input struct {
	a      bool
	b      bool
	c      []bool
	expect bool
	output output
}

func TestWalkStruct(t *testing.T) {
	var a input
	var b input
	if reflect.TypeOf(a) == reflect.TypeOf(b) {
		t.Logf("typ of a == typ of b")
	}

	Walk(reflect.TypeOf(input{}), true, func(s reflect.Type, sf reflect.StructField) (stop bool) {
		t.Logf("typ:%v sf:%v", s, sf)
		return false
	})
}

func TestWalkBool(t *testing.T) {
	var a bool
	var b bool
	if reflect.TypeOf(a) == reflect.TypeOf(b) {
		t.Logf("typ of a == typ of b")
	}
	Walk(reflect.TypeOf(bool(false)), false, func(s reflect.Type, sf reflect.StructField) (stop bool) {
		t.Logf("typ:%v sf:%v", s, sf)
		return false
	})
}
