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
	OnHandshakeHandler
	OnOpenHandler
	OnMsgReadHandler   //Block
	OnMsgHandleHandler //Unblock
	OnCloseHandler
	OnErrorHandler
}

func NewServerFunc(onHandshake OnHandshakeHandler,
	onOpen OnOpenHandler,
	onMsgRead OnMsgReadHandler,
	onMsgHandle OnMsgHandleHandler,
	onClose OnCloseHandler,
	onError OnErrorHandler) *Server {
	return &Server{
		onHandshakeHandler: object.RequireNonNullElse(onHandshake, NopOnHandshakeHandler).(OnHandshakeHandler),
		onOpenHandler:      object.RequireNonNullElse(onOpen, NopOnOpenHandler).(OnOpenHandler),
		onMsgReadHandler:   object.RequireNonNullElse(onMsgRead, NopOnMsgReadHandler).(OnMsgReadHandler),
		onMsgHandleHandler: object.RequireNonNullElse(onMsgHandle, NopOnMsgHandleHandler).(OnMsgHandleHandler),
		onCloseHandler:     object.RequireNonNullElse(onClose, NopOnCloseHandler).(OnCloseHandler),
		onErrorHandler:     object.RequireNonNullElse(onError, NopOnErrorHandler).(OnErrorHandler),
	}
}
func NewServer(h ServerHandler) *Server {
	return NewServerFunc(h, h, h, h, h, h)
}

type Server struct {
	upgrader           websocket.Upgrader // use default options
	onHandshakeHandler OnHandshakeHandler
	onOpenHandler      OnOpenHandler
	onMsgReadHandler   OnMsgReadHandler
	onMsgHandleHandler OnMsgHandleHandler
	onCloseHandler     OnCloseHandler
	onErrorHandler     OnErrorHandler

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
	ConnState func(*WebSocketConn, ConnState)
}

func (srv *Server) CheckError(conn WebSocketReadWriteCloser, err error) error {
	if err == nil {
		return nil
	}
	return srv.onErrorHandler.OnError(conn, err)
}

// OnHandshake takes over the http handler
func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	if srv.shuttingDown() {
		return ErrServerClosed
	}
	// transfer http to websocket
	srv.upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := srv.upgrader.Upgrade(w, r, nil)
	if srv.CheckError(nil, err) != nil {
		return err
	}
	defer ws.Close()
	ctx := context.WithValue(context.Background(), ServerContextKey, srv)
	// Handle HTTP Handshake
	err = srv.onHandshakeHandler.OnHandshake(w, r)
	if srv.CheckError(nil, err) != nil {
		return err
	}
	// takeover the connect
	c := srv.newConn(ws)
	// Handle websocket On
	err = srv.onOpenHandler.OnOpen(c.rwc)
	if srv.CheckError(c.rwc, err) != nil {
		c.close()
		return err
	}
	c.setState(c.rwc, StateNew) // before Serve can return

	return c.serve(ctx)
}

// Create new connection from rwc.
func (srv *Server) newConn(wc *websocket.Conn) *conn {
	c := &conn{
		server: srv,
		rwc: &WebSocketConn{
			Conn: wc,
		},
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
