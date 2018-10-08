package tcp

import (
	"bufio"
	"github.com/searKing/golib/util/object"
	"sync"
)

type TCPConn struct {
	*bufio.ReadWriter
	muRead  sync.Mutex
	muWrite sync.Mutex
}

func NewTCPConn(rw *bufio.ReadWriter) *TCPConn {
	object.RequireNonNil(rw)
	return &TCPConn{
		ReadWriter: rw,
	}
}
