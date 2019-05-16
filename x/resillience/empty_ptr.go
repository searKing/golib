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

// BackgroundPtr returns a non-nil, empty Ptr.
func BackgroundPtr() Ptr {
	return backgroundPtr
}

// TODO returns a non-nil, empty Ptr. Code should use context.TODO when
// it's unclear which Ptr to use or it is not yet available .
func TODOPtr() Ptr {
	return todoPtr
}
