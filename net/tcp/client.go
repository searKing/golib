package tcp

type ClientHandler interface {
	OnOpenHandler
	OnMsgReadHandler
	OnMsgHandleHandler
	OnCloseHandler
	OnErrorHandler
}
