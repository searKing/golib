package default_

import "reflect"

type convertFunc func(v reflect.Value, tag reflect.StructTag) error
