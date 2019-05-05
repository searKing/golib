package resilience

type emptyPtr int

func (r *emptyPtr) Value() interface{} {
	return nil
}

func (r *emptyPtr) Ready() error {
	return nil
}

func (r *emptyPtr) Close() {
	return
}

var (
	backgroundPtr = new(emptyPtr)
	todoPtr       = new(emptyPtr)
)

// BackgroundPtr returns a non-nil, empty Context. It is never canceled, has no
// values, and has no deadline. It is typically used by the main function,
// initialization, and tests, and as the top-level Context for incoming
// requests.
func BackgroundPtr() Ptr {
	return backgroundPtr
}

// TODO returns a non-nil, empty Context. Code should use context.TODO when
// it's unclear which Context to use or it is not yet available (because the
// surrounding function has not yet been extended to accept a Context
// parameter).
func TODOPtr() Ptr {
	return todoPtr
}
