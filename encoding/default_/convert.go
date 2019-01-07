package default_

type convOpts struct{}

// Convert wrapper of convertState
func Convert(v interface{}) error {
	e := newConvertState()
	err := e.convert(v, convOpts{})
	if err != nil {
		return err
	}

	e.Reset()
	convertStatePool.Put(e)
	return nil
}

// Marshaler is the interface implemented by types that
// can marshal themselves into valid JSON.
type Converter interface {
	ConvertDefault() error
}
