package tcp

import (
	"context"
	"github.com/searKing/golib/time/delay"
	"github.com/searKing/golib/util/object"
	"go.uber.org/atomic"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type ServerHandler interface {
	OnMsgReadHandler
	OnMsgHandleHandler
}

func NewServerFunc(readMsgHandler OnMsgReadHandler, handleMsgHandler OnMsgHandleHandler) *Server {
	return &Server{
		ReadMsgHandler:   object.RequireNonNullElse(readMsgHandler, NopReadMsgHandler).(OnMsgReadHandler),
		HandleMsgHandler: object.RequireNonNullElse(handleMsgHandler, NopMsgHandlerFunc).(OnMsgHandleHandler),
	}
}
func NewServer(h ServerHandler) *Server {
	return NewServerFunc(h, h)
}

var NopReadMsgHandler = OnMsgReadHandlerFunc(func(b io.Reader) (msg interface{}, err error) { return nil, nil })
var NopMsgHandlerFunc = OnMsgHandleHandlerFunc(func(b io.Writer, msg interface{}) error { return nil })

type Server struct {
	Addr             string // TCP address to listen on, ":tcp" if empty
	ReadMsgHandler   OnMsgReadHandler
	HandleMsgHandler OnMsgHandleHandler

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	MaxBytes     int

	ErrorLog *log.Logger

	mu         sync.Mutex
	listeners  map[*net.Listener]struct{}
	activeConn map[*conn]struct{}
	doneChan   chan struct{}
	onShutdown []func()

	// server state
	inShutdown atomic.Bool

	// ConnState specifies an optional callback function that is
	// called when a client connection changes state. See the
	// ConnState type and associated constants for details.
	ConnState func(net.Conn, ConnState)
}

func (srv *Server) ListenAndServe() error {
	if srv.shuttingDown() {
		return ErrServerClosed
	}
	addr := srv.Addr
	if addr == "" {
		addr = ":tcp"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return srv.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)})
}

func (srv *Server) Serve(l net.Listener) error {
	l = &onceCloseListener{Listener: l}
	defer l.Close()

	var tempDelay = delay.NewDefaultDelay() // how long to sleep on accept failure
	ctx := context.WithValue(context.Background(), ServerContextKey, srv)
	for {
		rw, e := l.Accept()
		if e != nil {
			// return if server is cancaled, means normally close
			select {
			case <-srv.getDoneChan():
				return ErrServerClosed
			default:
			}
			// retry if it's recoverable
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				tempDelay.Update()
				srv.logf("http: Accept error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay.Duration())
				continue
			}
			// return otherwise
			return e
		}
		tempDelay.Reset()

		// takeover the connect
		c := srv.newConn(rw)
		c.setState(c.rwc, StateNew) // before Serve can return
		go c.serve(ctx)
	}
}

func (s *Server) trackConn(c *conn, add bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.activeConn == nil {
		s.activeConn = make(map[*conn]struct{})
	}
	if add {
		s.activeConn[c] = struct{}{}
	} else {
		delete(s.activeConn, c)
	}
}

// Create new connection from rwc.
func (srv *Server) newConn(rwc net.Conn) *conn {
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

func ListenAndServe(addr string, readMsg OnMsgReadHandler, handleMsg OnMsgHandleHandler) error {
	server := &Server{Addr: addr, ReadMsgHandler: readMsg, HandleMsgHandler: handleMsg}
	return server.ListenAndServe()
}
