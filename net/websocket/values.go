package websocket

import "errors"

// ErrServerClosed is returned by the Server's Serve and ListenAndServe
// methods after a call to Shutdown or Close.
var ErrServerClosed = errors.New("websocket: Server closed")
var ErrNotFound = errors.New("websocket: Server not found")
