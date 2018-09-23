package slice

import (
	"github.com/searKing/golib/util/object"
	"sync"
)

// ForEachFunc Performs an action for each element of this slice.
// <p>The behavior of this operation is explicitly nondeterministic.
// For parallel slice pipelines, this operation does <em>not</em>
// guarantee to respect the encounter order of the slice, as doing so
// would sacrifice the benefit of parallelism.  For any given element, the
// action may be performed at whatever time and in whatever thread the
// library chooses.  If the action accesses shared state, it is
// responsible for providing the required synchronization.
func ForEachFunc(s interface{}, f func(interface{}), ifStringAsRune ...bool) {
	forEachFunc(Of(s, ifStringAsRune...), f)
}

// forEachFunc is the same as ForEachFunc
func forEachFunc(s []interface{}, f func(interface{})) {
	object.RequireNonNil(s, "forEachFunc called on nil slice")
	object.RequireNonNil(s, "forEachFunc called on nil callfn")
	var wg sync.WaitGroup
	for _, r := range s {
		wg.Add(1)
		go func(rr interface{}) {
			f(rr)
			wg.Done()
		}(r)
	}
	wg.Wait()
	return
}
