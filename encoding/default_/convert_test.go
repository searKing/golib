package default_

import (
	"reflect"
	"testing"
)

type inputType struct {
	Name        Name              `default:"Alice"`
	Age         int               `default:"10"`
	IntArray    []int             `default:"[1,2,3]"`
	StringArray []string          `default:"[\"stdout\",\"./logs\"]"`
	Map         map[string]string `default:"{\"name\": \"Alice\", \"age\": 18}"`
}
type Name string

func (thiz *Name) ConvertDefault() error {
	if *thiz == "" {
		*thiz = "Bob"
	}
	return nil
}
func TestConvert(t *testing.T) {
	i := &inputType{}
	expect := &inputType{
		Name:        "Alice",
		Age:         10,
		IntArray:    []int{1, 2, 3},
		StringArray: []string{"stdout", "./logs"},
		Map:         map[string]string{"name": "Alice", "age": "18"},
	}
	err := Convert(i)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(i, expect) {
		t.Errorf("expect\n[\n%v\n]\nactual[\n%v\n]", expect, i)
	}
}
