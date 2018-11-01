package lang

type NullPointerException struct {
	ThrowableInterface
}

func NewNullPointerException(message string) ThrowableInterface {
	return &NullPointerException{}

}
