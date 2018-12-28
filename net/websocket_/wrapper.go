package websocket_

import (
	"github.com/gorilla/websocket"
	"github.com/searKing/golib/util/object"
	"gitlab.hobot.cc/haixin.chen/go-mvc/transport"
	"sync"
)

// make websocket concurrent safe
// see https://godoc.org/github.com/gorilla/websocket#hdr-Concurrency
type WebSocketConn struct {
	*websocket.Conn
	muRead  sync.Mutex
	muWrite sync.Mutex
}

func NewWebSocketConn(rw *websocket.Conn) transport.ReadWriteCloser {
	object.RequireNonNil(rw)
	return &WebSocketConn{
		Conn: rw,
	}
}
func (c *WebSocketConn) ReadMessage() (messageType int, p []byte, err error) {
	c.muRead.Lock()
	defer c.muRead.Unlock()
	return c.Conn.ReadMessage()
}
func (c *WebSocketConn) WriteJSON(v interface{}) error {
	c.muWrite.Lock()
	defer c.muWrite.Unlock()
	return c.Conn.WriteJSON(v)
}
func (c *WebSocketConn) Close() error {
	c.muWrite.Lock()
	defer c.muWrite.Unlock()
	return c.Conn.Close()
}
