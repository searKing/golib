package tcp

import (
	"bufio"
	"github.com/searKing/golib/util/object"
	"sync"
)

type Handler interface {
	ReadMsgHandler
	HandleMsgHandler
}
type handlerFunc struct {
	read   ReadMsgHandler
	handle HandleMsgHandler
}

func HandlerFunc(read func(b *bufio.Reader) (msg interface{}, err error), handle func(b *bufio.Writer, msg interface{}) error) Handler {
	return &handlerFunc{
		read:   ReadMsgHandlerFunc(read),
		handle: HandleMsgHandlerFunc(handle),
	}
}
func (f handlerFunc) ReadMsg(b *bufio.Reader) (msg interface{}, err error) {
	return f.read.ReadMsg(b)
}
func (f handlerFunc) HandleMsg(b *bufio.Writer, msg interface{}) error {
	return f.handle.HandleMsg(b, msg)
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

func (mux *ServeMux) ReadMsg(b *bufio.Reader) (req interface{}, err error) {
	return mux.msgHandler.ReadMsg(b)
}

func (mux *ServeMux) HandleMsg(b *bufio.Writer, msg interface{}) error {
	return mux.msgHandler.HandleMsg(b, msg)
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
