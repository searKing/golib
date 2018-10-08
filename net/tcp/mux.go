package tcp

import (
	"bufio"
	"github.com/searKing/golib/util/object"
	"io"
	"sync"
)

type Handler interface {
	OnMsgReadHandler
	OnMsgHandleHandler
}
type handlerFunc struct {
	read   OnMsgReadHandler
	handle OnMsgHandleHandler
}

func HandlerFunc(read func(b io.Reader) (msg interface{}, err error), handle func(b io.Writer, msg interface{}) error) Handler {
	return &handlerFunc{
		read:   OnMsgReadHandlerFunc(read),
		handle: OnMsgHandleHandlerFunc(handle),
	}
}
func (f handlerFunc) OnMsgRead(b io.Reader) (msg interface{}, err error) {
	return f.read.OnMsgRead(b)
}
func (f handlerFunc) OnMsgHandle(b io.Writer, msg interface{}) error {
	return f.handle.OnMsgHandle(b, msg)
}

type ServeMux struct {
	mu         sync.RWMutex
	msgHandler Handler
}

// NewServeMux allocates and returns a new ServeMux.
func NewServeMux() *ServeMux {
	return &ServeMux{}
}

// DefaultServeMux is the default ServeMux used by Serve.
var DefaultServeMux = &defaultServeMux

var defaultServeMux ServeMux

func (mux *ServeMux) OnMsgRead(b io.Reader) (req interface{}, err error) {
	return mux.msgHandler.OnMsgRead(b)
}

func (mux *ServeMux) OnMsgHandle(b io.Writer, msg interface{}) error {
	return mux.msgHandler.OnMsgHandle(b, msg)
}
func (mux *ServeMux) Handle(handler Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()
	object.RequireNonNil(handler, "tcp: nil handler")
	mux.msgHandler = handler
}
func (mux *ServeMux) handle() Handler {
	mux.mu.RLock()
	defer mux.mu.RUnlock()
	if mux.msgHandler == nil {
		return NotFoundHandler()
	}
	return mux.msgHandler
}
func NotFoundHandler() Handler { return &NotFound{} }

// NotFoundHandler returns a simple request handler
// that replies to each request with a ``404 page not found'' reply.
type NotFound struct {
	Handler
}

func (notfound *NotFound) ReadMsg(b *bufio.Reader) (msg interface{}, err error) {
	return nil, ErrNotFound
}
func (notfound *NotFound) HandleMsg(b *bufio.Writer, msg interface{}) error {
	return ErrServerClosed
}
