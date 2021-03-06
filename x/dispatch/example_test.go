package dispatch_test

import (
	"context"
	"errors"
	"github.com/searKing/golib/x/dispatch"
	"log"
)

type DispatchMsg struct {
	data int
}

func ExampleDispatch() {
	var conn chan DispatchMsg
	dispatch.NewDispatch(
		dispatch.ReaderFunc(func(ctx context.Context) (interface{}, error) {
			return ReadMessage(conn)
		}),
		dispatch.HandlerFunc(func(ctx context.Context, msg interface{}) error {
			m := msg.(*DispatchMsg)
			return HandleMessage(m)
		})).Start()
}
func ExampleDispatcher_Join() {
	var conn chan DispatchMsg

	workflow := dispatch.NewDispatch(
		dispatch.ReaderFunc(func(ctx context.Context) (interface{}, error) {
			return ReadMessage(conn)
		}),
		dispatch.HandlerFunc(func(ctx context.Context, msg interface{}) error {
			m := msg.(*DispatchMsg)
			return HandleMessage(m)
		})).Joinable()

	go func() {
		workflow.Start()
	}()
	const cnt = 10
	for i := 0; i < cnt; i++ {
		conn <- DispatchMsg{data: i}
	}
	workflow.Join()
}
func ExampleDispatcher_Context() {
	var conn chan DispatchMsg

	workflow := dispatch.NewDispatch(
		dispatch.ReaderFunc(func(ctx context.Context) (interface{}, error) {
			return ReadMessage(conn)
		}),
		dispatch.HandlerFunc(func(ctx context.Context, msg interface{}) error {
			m := msg.(*DispatchMsg)
			return HandleMessage(m)
		})).Joinable()
	go func() {
		workflow.Start()
	}()
	workflow.Context().Done()
	workflow.Join()
}

func ReadMessage(c <-chan DispatchMsg) (interface{}, error) {
	var msg DispatchMsg
	var ok bool

	if msg, ok = <-c; ok {
		log.Println("Recv From Channel Failed")
		return nil, errors.New("Recv From Channel Failed")
	}
	log.Printf("Recv From Channel Success: %v\n", msg.data)
	return &msg, nil
}

// just print what's received
func HandleMessage(msg *DispatchMsg) error {
	if msg == nil {
		log.Println("Handle From Channel Failed")
		return errors.New("Handle From Channel Failed")
	}
	log.Printf("Handle From Channel Success: %v\n", msg.data)
	return nil
}
