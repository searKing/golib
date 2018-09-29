package slice

import (
	"github.com/searKing/golib/util/object"
)

// MaxFunc returns the maximum element of this stream according to the provided.
func MaxFunc(s interface{}, f func(interface{}, interface{}) int) interface{} {
	return normalizeElem(minFunc(Of(s), f), s)

}

// maxFunc is the same as MaxFunc
func maxFunc(s []interface{}, f func(interface{}, interface{}) int) interface{} {
	object.RequireNonNil(s, "maxFunc called on nil slice")
	object.RequireNonNil(f, "maxFunc called on nil callfn")

	return ReduceFunc(s, func(left, right interface{}) interface{} {
		if f(left, right) > 0 {
			return left
		}
		return right
	})
}
