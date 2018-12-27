package reflect_

import (
	"testing"
	"time"
)

type inputObject struct {
	a      interface{}
	expect bool
}

func TestValueIsNilObject(t *testing.T) {
	var nilTime *time.Time
	ins := []inputObject{
		{
			a:      nil,
			expect: true,
		},
		{
			a:      true,
			expect: false,
		},
		{
			a:      0,
			expect: false,
		},
		{
			a:      "",
			expect: false,
		},
		{
			a:      time.Now(),
			expect: false,
		},
		{
			a:      nilTime,
			expect: true,
		},
	}
	for idx, in := range ins {
		if IsNilObject(in.a) != in.expect {
			t.Errorf("#%d expect %t", idx, in.expect)
		}
	}
}
