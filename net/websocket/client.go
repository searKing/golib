package websocket

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/searKing/golib/util/object"
	"net/http"
	"net/url"
)

type ClientHandler interface {
	OnHTTPResponseHandler
	OnOpenHandler
	OnMsgReadHandler
	OnMsgHandleHandler
	OnCloseHandler
	OnErrorHandler
}
type Client struct {
	*Server
	httpRespHandler OnHTTPResponseHandler
	url             url.URL
}

func NewClientFunc(onHTTPRespHandler OnHTTPResponseHandler,
	onOpenHandler OnOpenHandler,
	onMsgReadHandler OnMsgReadHandler,
	onMsgHandleHandler OnMsgHandleHandler,
	onCloseHandler OnCloseHandler,
	onErrorHandler OnErrorHandler) *Client {
	return &Client{
		Server:          NewServerFunc(nil, onOpenHandler, onMsgReadHandler, onMsgHandleHandler, onCloseHandler, onErrorHandler),
		httpRespHandler: object.RequireNonNullElse(onHTTPRespHandler, NopOnHTTPResponseHandler).(OnHTTPResponseHandler),
	}
}
func NewClient(h ClientHandler) *Client {
	return NewClientFunc(h, h, h, h, h, h)
}

// OnHandshake takes over the http handler
func (cli *Client) ServeHTTP(requestHeader http.Header) error {
	if cli.shuttingDown() {
		return ErrServerClosed
	}
	// transfer http to websocket
	ws, resp, err := websocket.DefaultDialer.Dial(cli.url.String(), requestHeader)
	if cli.Server.CheckError(nil, err) != nil {
		return err
	}
	// Handle HTTP Response
	err = cli.httpRespHandler.OnHTTPResponse(resp)
	if cli.Server.CheckError(nil, err) != nil {
		return err
	}

	defer ws.Close()
	ctx := context.WithValue(context.Background(), ClientContextKey, cli)

	// takeover the connect
	c := cli.Server.newConn(ws)
	// Handle websocket On
	err = cli.Server.onOpenHandler.OnOpen(c.rwc)
	if err = cli.Server.CheckError(c.rwc, err); err != nil {
		c.close()
		return err
	}
	c.setState(c.rwc, StateNew) // before Serve can return

	return c.serve(ctx)
}
