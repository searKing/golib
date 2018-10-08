package tcp

import "bufio"

type OnMsgReadHandler interface {
	OnMsgRead(b *bufio.Reader) (msg interface{}, err error)
}
type OnMsgReadHandlerFunc func(b *bufio.Reader) (msg interface{}, err error)

func (f OnMsgReadHandlerFunc) OnMsgRead(b *bufio.Reader) (msg interface{}, err error) {
	return f(b)
}

type OnMsgHandleHandler interface {
	OnMsgHandle(b *bufio.Writer, msg interface{}) error
}
type OnMsgHandleHandlerFunc func(b *bufio.Writer, msg interface{}) error

func (f OnMsgHandleHandlerFunc) OnMsgHandle(b *bufio.Writer, msg interface{}) error {
	return f(b, msg)
}
