package default_

import (
	"github.com/searKing/golib/encoding/internal/tag"
	"github.com/searKing/golib/reflect_"
	"gopkg.in/yaml.v2"
	"reflect"
)

const TagDefault = "default"

// Convert wrapper of convertState
func Convert(val interface{}) error {
	return tag.Tag(val, func(val reflect.Value, tag reflect.StructTag) error {
		if !reflect_.IsEmptyValue(val) {
			return nil
		}
		defaultTag, ok := tag.Lookup(TagDefault)
		if !ok {
			return nil
		}
		return yaml.Unmarshal([]byte(defaultTag), val.Addr().Interface())
	})
}

// Marshaler is the interface implemented by types that
// can marshal themselves into valid JSON.
type Converter interface {
	ConvertDefault() error
}
