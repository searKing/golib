package slice

import (
	"github.com/searKing/golib/util/object"
	"sync"
)

// FindAnyFunc returns an {@link Optional} describing some element of the stream, or an
// empty {@code Optional} if the stream is empty.
func FindAnyFunc(s interface{}, f func(interface{}) bool, ifStringAsRune ...bool) interface{} {
	return anyMatchFunc(Of(s, ifStringAsRune...), f, true)
}

// findAnyFunc is the same as FindAnyFunc.
func findAnyFunc(s []interface{}, f func(interface{}) bool, truth bool) interface{} {
	object.RequireNonNil(s, "findFirstFunc called on nil slice")
	object.RequireNonNil(f, "findFirstFunc called on nil callfn")
	var findc chan interface{}
	findc = make(chan interface{})
	defer close(findc)
	var mu sync.Mutex
	var wg sync.WaitGroup
	var found bool
	hasFound := func() bool {
		mu.Lock()
		defer mu.Unlock()
		return found
	}
	for _, r := range s {
		if hasFound() {
			break
		}

		wg.Add(1)
		go func(rr interface{}) {
			defer wg.Done()
			foundYet := func() bool {
				mu.Lock()
				defer mu.Unlock()
				return found
			}()
			if foundYet {
				return
			}
			if f(rr) == truth {
				mu.Lock()
				defer mu.Unlock()
				if found {
					return
				}
				found = true
				findc <- rr
				return
			}
		}(r)
	}
	go func() {
		defer close(findc)
		wg.Done()
	}()
	out, ok := <-findc
	if ok {
		return out
	}
	return nil
}
