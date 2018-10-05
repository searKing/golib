package websocket

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/searKing/golib/util/object"
	"go.uber.org/atomic"
	"log"
	"net/http"
	"sync"
	"time"
)

type ServerHandler interface {
	HTTPHandler
	ReadMsgHandler
	HandleMsgHandler
}
type HTTPHandler interface {
	ServeHTTP(http.ResponseWriter, *http.Request) error
}
type HTTPHandlerFunc func(http.ResponseWriter, *http.Request) error

func (f HTTPHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return f(w, r)
}

type ReadMsgHandler interface {
	ReadMsg(conn WebSocketReadWriteCloser) (msg interface{}, err error)
}
type ReadMsgHandlerFunc func(conn WebSocketReadWriteCloser) (msg interface{}, err error)

func (f ReadMsgHandlerFunc) ReadMsg(conn WebSocketReadWriteCloser) (msg interface{}, err error) {
	return f(conn)
}

type HandleMsgHandler interface {
	HandleMsg(conn WebSocketReadWriteCloser, msg interface{}) error
}
type HandleMsgHandlerFunc func(conn WebSocketReadWriteCloser, msg interface{}) error

func (f HandleMsgHandlerFunc) HandleMsg(conn WebSocketReadWriteCloser, msg interface{}) error {
	return f(conn, msg)
}

var NopHTTPHandler = HTTPHandlerFunc(func(http.ResponseWriter, *http.Request) error { return nil })
var NopReadMsgHandler = ReadMsgHandlerFunc(func(conn WebSocketReadWriteCloser) (msg interface{}, err error) { return nil, nil })
var NopMsgHandlerFunc = HandleMsgHandlerFunc(func(conn WebSocketReadWriteCloser, msg interface{}) error { return nil })

func NewServerFunc(httpHandler HTTPHandler, readMsgHandler ReadMsgHandler, handleMsgHandler HandleMsgHandler) *Server {
	return &Server{
		HTTPHandler:      object.RequireNonNullElse(httpHandler, NopHTTPHandler).(HTTPHandler),
		ReadMsgHandler:   object.RequireNonNullElse(readMsgHandler, NopReadMsgHandler).(ReadMsgHandler),
		HandleMsgHandler: object.RequireNonNullElse(handleMsgHandler, NopMsgHandlerFunc).(HandleMsgHandler),
	}
}
func NewServer(h ServerHandler) *Server {
	return NewServerFunc(h, h, h)
}

type Server struct {
	upgrader         websocket.Upgrader // use default options
	HTTPHandler      HTTPHandler
	ReadMsgHandler   ReadMsgHandler
	HandleMsgHandler HandleMsgHandler

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	MaxBytes     int

	ErrorLog *log.Logger

	mu         sync.Mutex
	activeConn *conn
	onShutdown []func()

	// server state
	disableKeepAlives atomic.Bool // accessed atomically.
	inShutdown        atomic.Bool
	// ConnState specifies an optional callback function that is
	// called when a client connection changes state. See the
	// ConnState type and associated constants for details.
	ConnState func(*websocket.Conn, ConnState)
}

// ServeHTTP takes over the http handler
func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	if srv.shuttingDown() {
		return ErrServerClosed
	}
	// transfer http to websocket
	srv.upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := srv.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer ws.Close()
	ctx := context.WithValue(context.Background(), ServerContextKey, srv)
	// Handle HTTP Handshake
	err = srv.HTTPHandler.ServeHTTP(w, r)
	if err != nil {
		return err
	}
	// takeover the connect
	c := srv.newConn(ws)
	c.setState(c.rwc, StateNew) // before Serve can return
	c.serve(ctx)
	return nil
}

// Create new connection from rwc.
func (srv *Server) newConn(rwc *websocket.Conn) *conn {
	c := &conn{
		server: srv,
		rwc:    rwc,
	}
	return c
}
func (s *Server) logf(format string, args ...interface{}) {
	if s.ErrorLog != nil {
		s.ErrorLog.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}
