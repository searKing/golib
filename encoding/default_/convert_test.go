package default_

import (
	"testing"
)

type inputType struct {
	Name Name `default:"Alice"`
	Age  int  `default:"10"`
}
type Name string

func (thiz *Name) ConvertDefault() error {
	*thiz = "Bob"
	return nil
}
func TestConvert(t *testing.T) {
	i := &inputType{}
	err := Convert(i)
	if err != nil {
		t.Error(err)
	}
	t.Logf("inputType = %v\n", i)
}
