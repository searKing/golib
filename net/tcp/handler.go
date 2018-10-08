package tcp

import (
	"io"
	"net"
)

type OnOpenHandler interface {
	OnOpen(conn net.Conn) error
}
type OnOpenHandlerFunc func(conn net.Conn) error

func (f OnOpenHandlerFunc) OnOpen(conn net.Conn) error { return f(conn) }

type OnMsgReadHandler interface {
	OnMsgRead(b io.Reader) (msg interface{}, err error)
}
type OnMsgReadHandlerFunc func(b io.Reader) (msg interface{}, err error)

func (f OnMsgReadHandlerFunc) OnMsgRead(b io.Reader) (msg interface{}, err error) { return f(b) }

type OnMsgHandleHandler interface {
	OnMsgHandle(b io.Writer, msg interface{}) error
}
type OnMsgHandleHandlerFunc func(b io.Writer, msg interface{}) error

func (f OnMsgHandleHandlerFunc) OnMsgHandle(b io.Writer, msg interface{}) error { return f(b, msg) }

type OnCloseHandler interface {
	OnClose(w io.Writer, r io.Reader) error
}
type OnCloseHandlerFunc func(w io.Writer, r io.Reader) error

func (f OnCloseHandlerFunc) OnClose(w io.Writer, r io.Reader) error { return f(w, r) }

type OnErrorHandler interface {
	OnError(w io.Writer, r io.Reader, err error) error
}
type OnErrorHandlerFunc func(w io.Writer, r io.Reader, err error) error

func (f OnErrorHandlerFunc) OnError(w io.Writer, r io.Reader, err error) error { return f(w, r, err) }
