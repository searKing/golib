package tcp

import (
	"bufio"
	"context"
	"errors"
	"github.com/searKing/dispatch"
	"io"
	"net"
	"runtime"
	"sync/atomic"
	"time"
)

// maxInt64 is the effective "infinite" value for the Server and
// Transport's byte-limiting readers.
const maxInt64 = 1<<63 - 1

// aLongTimeAgo is a non-zero time, far in the past, used for
// immediate cancelation of network operations.
var aLongTimeAgo = time.Unix(1, 0)

// A conn represents the server side of an HTTP connection.
type conn struct {
	// server is the server on which the connection arrived.
	// Immutable; never nil.
	server *Server

	// cancelCtx cancels the connection-level context.
	cancelCtx context.CancelFunc

	// rwc is the underlying network connection.
	// This is never wrapped by other types and is the value given out
	// to CloseNotifier callers. It is usually of type *net.TCPConn or
	// *tls.Conn.
	rwc net.Conn

	// remoteAddr is rwc.RemoteAddr().String(). It is not populated synchronously
	// inside the Listener's Accept goroutine, as some implementations block.
	// It is populated immediately inside the (*conn).serve goroutine.
	// This is the value of a HandleMsgHandler's (*Request).RemoteAddr.
	remoteAddr string

	// werr is set to the first write error to rwc.
	// It is set via checkConnErrorWriter{w}, where bufw writes.
	werr error

	// r is bufr's read source. It's a wrapper around rwc that provides
	// io.LimitedReader-style limiting (while reading request headers)
	// and functionality to support CloseNotifier. See *connReader docs.
	r *connReader

	// bufr reads from r.
	bufr *bufio.Reader

	// bufw writes to checkConnErrorWriter{c}, which populates werr on error.
	bufw *bufio.Writer

	curState struct{ atomic uint64 } // packed (unixtime<<8|uint8(ConnState))
}

func (c *conn) finalFlush() {
	if c.bufr != nil {
		// Steal the bufio.ReadMsgHandler (~4KB worth of memory) and its associated
		// reader for a future connection.
		putBufioReader(c.bufr)
		c.bufr = nil
	}

	if c.bufw != nil {
		c.bufw.Flush()
		// Steal the bufio.Writer (~4KB worth of memory) and its associated
		// writer for a future connection.
		putBufioWriter(c.bufw)
		c.bufw = nil
	}
}

// Close the connection.
func (c *conn) close() {
	c.finalFlush()
	c.rwc.Close()
}

// rstAvoidanceDelay is the amount of time we sleep after closing the
// write side of a TCP connection before closing the entire socket.
// By sleeping, we increase the chances that the client sees our FIN
// and processes its final data before they process the subsequent RST
// from closing a connection with known unread data.
// This RST seems to occur mostly on BSD systems. (And Windows?)
// This timeout is somewhat arbitrary (~latency around the planet).
const rstAvoidanceDelay = 500 * time.Millisecond

type closeWriter interface {
	CloseWrite() error
}

var _ closeWriter = (*net.TCPConn)(nil)

// closeWrite flushes any outstanding data and sends a FIN packet (if
// client is connected via TCP), signalling that we're done. We then
// pause for a bit, hoping the client processes it before any
// subsequent RST.
//
// See https://golang.org/issue/3595
func (c *conn) closeWriteAndWait() {
	c.finalFlush()
	if tcp, ok := c.rwc.(closeWriter); ok {
		tcp.CloseWrite()
	}
	time.Sleep(rstAvoidanceDelay)
}

func (c *conn) setState(nc net.Conn, state ConnState) {
	srv := c.server
	if state > 0xff || state < 0 {
		panic("internal error")
	}
	packedState := uint64(time.Now().Unix()<<8) | uint64(state)
	atomic.StoreUint64(&c.curState.atomic, packedState)
	if hook := srv.ConnState; hook != nil {
		hook(nc, state)
	}
}

func (c *conn) getState() (state ConnState, unixSec int64) {
	packedState := atomic.LoadUint64(&c.curState.atomic)
	return ConnState(packedState & 0xff), int64(packedState >> 8)
}

// ErrAbortHandler is a sentinel panic value to abort a handler.
// While any panic from ServeHTTP aborts the response to the client,
// panicking with ErrAbortHandler also suppresses logging of a stack
// trace to the server's error log.
var ErrAbortHandler = errors.New("net/tcp: abort HandleMsgHandler")
var errTooLarge = errors.New("tcp: read too large")

// isCommonNetReadError reports whether err is a common error
// encountered during reading a request off the network when the
// client has gone away or had its read fail somehow. This is used to
// determine which logs are interesting enough to log about.
func isCommonNetReadError(err error) bool {
	if err == io.EOF {
		return true
	}
	if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
		return true
	}
	if oe, ok := err.(*net.OpError); ok && oe.Op == "read" {
		return true
	}
	return false
}

// ReadMsgHandler next request from connection.
func (c *conn) readRequest(ctx context.Context) (req interface{}, err error) {

	var (
		wholeReqDeadline time.Time // or zero if none
		hdrDeadline      time.Time // or zero if none
	)
	t0 := time.Now()
	if d := c.server.readTimeout(); d != 0 {
		hdrDeadline = t0.Add(d)
	}
	if d := c.server.ReadTimeout; d != 0 {
		wholeReqDeadline = t0.Add(d)
	}
	c.rwc.SetReadDeadline(hdrDeadline)
	if d := c.server.WriteTimeout; d != 0 {
		defer func() {
			c.rwc.SetWriteDeadline(time.Now().Add(d))
		}()
	}

	c.r.setReadLimit(c.server.initialReadLimitSize())
	req, err = c.server.ReadMsgHandler.ReadMsg(c.bufr)
	if err != nil {
		if c.r.hitReadLimit() {
			return nil, errTooLarge
		}
		return nil, err
	}

	c.r.setInfiniteReadLimit()

	// Adjust the read deadline if necessary.
	if !hdrDeadline.Equal(wholeReqDeadline) {
		c.rwc.SetReadDeadline(wholeReqDeadline)
	}

	return req, nil
}

// Serve a new connection.
func (c *conn) serve(ctx context.Context) {
	c.remoteAddr = c.rwc.RemoteAddr().String()
	ctx = context.WithValue(ctx, LocalAddrContextKey, c.rwc.LocalAddr())
	// handle close
	defer func() {
		if err := recover(); err != nil && err != ErrAbortHandler {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			c.server.logf("tcp: panic serving %v: %v\n%s", c.remoteAddr, err, buf)
		}
		c.close()
		c.setState(c.rwc, StateClosed)
	}()

	// TCP from here on.
	// cancel after this connection is handled
	ctx, cancelCtx := context.WithCancel(ctx)
	c.cancelCtx = cancelCtx
	defer cancelCtx()

	// wrap original conn itself with buffer
	c.r = &connReader{conn: c}
	c.bufr = newBufioReader(c.r)
	c.bufw = newBufioWriterSize(checkConnErrorWriter{c}, 4<<10)

	// read and handle the msg
	dispatch.NewDispatch(dispatch.ReaderFunc(func() (interface{}, error) {
		msg, err := c.readRequest(ctx)
		if c.r.remain != c.server.initialReadLimitSize() {
			// If we read any bytes off the wire, we're active.
			c.setState(c.rwc, StateActive)
		}
		if err != nil {
			if isCommonNetReadError(err) {
				return nil, err // don't reply
			}
			// otherwise, close socket's Write channel and wait for a monment
			c.closeWriteAndWait()
			return nil, err
		}
		return msg, nil
	}), dispatch.HandlerFunc(func(msg interface{}) error {
		return c.server.HandleMsgHandler.HandleMsg(c.bufw, msg)
	})).WithContext(ctx).Start()
	// after HandleMsgHandler, read all left data
	c.r.startBackgroundRead()
	finishConnect := func() {
		c.bufw.Flush()
		c.rwc.Close()
		c.r.abortPendingRead()
	}
	finishConnect()
}

// checkConnErrorWriter writes to c.rwc and records any write errors to c.werr.
// It only contains one field (and a pointer field at that), so it
// fits in an interface value without an extra allocation.
type checkConnErrorWriter struct {
	c *conn
}

func (w checkConnErrorWriter) Write(p []byte) (n int, err error) {
	n, err = w.c.rwc.Write(p)
	if err != nil && w.c.werr == nil {
		w.c.werr = err
		w.c.cancelCtx()
	}
	return
}