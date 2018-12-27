package reflect_

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

type inputValue struct {
	a      reflect.Value
	expect bool
}

func TestValueIsZeroValue(t *testing.T) {
	ins := []inputValue{
		{
			a:      reflect.ValueOf(nil),
			expect: true,
		},
		{
			a:      reflect.ValueOf(true),
			expect: false,
		},
		{
			a:      reflect.ValueOf(0),
			expect: true,
		},
		{
			a:      reflect.ValueOf(""),
			expect: true,
		},
		{
			a:      reflect.ValueOf(time.Now()),
			expect: false,
		},
		{
			a:      reflect.ValueOf(time.Time{}),
			expect: true,
		},
	}
	for idx, in := range ins {
		if IsZeroValue(in.a) != in.expect {
			t.Errorf("#%d expect %t", idx, in.expect)
		}
	}
}
func TestValueIsNilValue(t *testing.T) {
	var nilTime *time.Time
	ins := []inputValue{
		{
			a:      reflect.ValueOf(nil),
			expect: true,
		},
		{
			a:      reflect.ValueOf(true),
			expect: false,
		},
		{
			a:      reflect.ValueOf(0),
			expect: false,
		},
		{
			a:      reflect.ValueOf(""),
			expect: false,
		},
		{
			a:      reflect.ValueOf(time.Now()),
			expect: false,
		},
		{
			a:      reflect.ValueOf(nilTime),
			expect: true,
		},
	}
	for idx, in := range ins {
		if IsNilValue(in.a) != in.expect {
			t.Errorf("#%d expect %t", idx, in.expect)
		}
	}
}

type inputDumpValue struct {
	a      reflect.Value
	expect string
}

func TestTypeDumpValueInfoDFS(t *testing.T) {
	var nilError *json.SyntaxError
	ins := []inputDumpValue{
		{
			a:      reflect.ValueOf(nil),
			expect: `<invalid Value>`,
		},
		{
			a:      reflect.ValueOf(true),
			expect: `[bool: true]`,
		},
		{
			a:      reflect.ValueOf(0),
			expect: `[int: 0]`,
		},
		{
			a:      reflect.ValueOf("HelloWorld"),
			expect: `[string: HelloWorld]`,
		},
		{
			a: reflect.ValueOf(json.SyntaxError{}),
			expect: `[json.SyntaxError: {msg: Offset:0}]
	[string: ]
	[int64: 0]`,
		},
		{
			a:      reflect.ValueOf(nilError),
			expect: `[*json.SyntaxError: <nil>]`,
		},
	}
	for idx, in := range ins {
		info := DumpValueInfoDFS(in.a)
		if info != in.expect {
			t.Errorf("#%d expect\n[\n%s\n]\nactual[\n%s\n]", idx, in.expect, info)
		}
	}
}

func TestTypeDumpValueInfoBFS(t *testing.T) {
	var nilError *json.SyntaxError
	ins := []inputDumpValue{
		{
			a:      reflect.ValueOf(nil),
			expect: `<invalid Value>`,
		},
		{
			a:      reflect.ValueOf(true),
			expect: `[bool: true]`,
		},
		{
			a:      reflect.ValueOf(0),
			expect: `[int: 0]`,
		},
		{
			a:      reflect.ValueOf(""),
			expect: `[string: ]`,
		},
		{
			a: reflect.ValueOf(json.SyntaxError{}),
			expect: `[json.SyntaxError: {msg: Offset:0}]
	[string: ]
	[int64: 0]`,
		},
		{
			a:      reflect.ValueOf(nilError),
			expect: `[*json.SyntaxError: <nil>]`,
		},
	}
	for idx, in := range ins {
		info := DumpValueInfoBFS(in.a)
		if info != in.expect {
			t.Errorf("#%d expect\n[\n%s\n]\nactual[\n%s\n]", idx, in.expect, info)
		}
	}
}
