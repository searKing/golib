package tcp

import (
	"bufio"
	"github.com/searKing/golib/util/object"
	"io"
	"net"
	"sync"
)

type ServeMux struct {
	mu         sync.RWMutex
	msgHandler ServerHandler
}

// NewServeMux allocates and returns a new ServeMux.
func NewServeMux() *ServeMux {
	return &ServeMux{}
}

// DefaultServeMux is the default ServeMux used by Serve.
var DefaultServeMux = &defaultServeMux

var defaultServeMux ServeMux

func (mux *ServeMux) OnOpen(conn net.Conn) error {
	return mux.msgHandler.OnOpen(conn)
}

func (mux *ServeMux) OnMsgRead(r io.Reader) (req interface{}, err error) {
	return mux.msgHandler.OnMsgRead(r)
}

func (mux *ServeMux) OnMsgHandle(w io.Writer, msg interface{}) error {
	return mux.msgHandler.OnMsgHandle(w, msg)
}
func (mux *ServeMux) OnClose(w io.Writer, r io.Reader) error {
	return mux.msgHandler.OnClose(w, r)
}
func (mux *ServeMux) OnError(w io.Writer, r io.Reader, err error) error {
	return mux.msgHandler.OnError(w, r, err)
}
func (mux *ServeMux) Handle(handler ServerHandler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()
	object.RequireNonNil(handler, "tcp: nil handler")
	mux.msgHandler = handler
}
func (mux *ServeMux) handle() ServerHandler {
	mux.mu.RLock()
	defer mux.mu.RUnlock()
	if mux.msgHandler == nil {
		return NotFoundHandler()
	}
	return mux.msgHandler
}
func NotFoundHandler() ServerHandler { return &NotFound{} }

// NotFoundHandler returns a simple request handler
// that replies to each request with a ``404 page not found'' reply.
type NotFound struct {
	NopServer
}

func (notfound *NotFound) ReadMsg(b *bufio.Reader) (msg interface{}, err error) {
	return nil, ErrNotFound
}
func (notfound *NotFound) HandleMsg(b *bufio.Writer, msg interface{}) error {
	return ErrServerClosed
}
