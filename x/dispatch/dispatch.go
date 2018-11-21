package dispatch

import (
	"context"
	"sync"
)

type Reader interface {
	Read(ctx context.Context) (msg interface{}, err error)
}
type ReaderFunc func(ctx context.Context) (msg interface{}, err error)

func (f ReaderFunc) Read(ctx context.Context) (msg interface{}, err error) {
	return f(ctx)
}

type Handler interface {
	Handle(ctx context.Context, msg interface{}) error
}
type HandlerFunc func(ctx context.Context, msg interface{}) error

func (f HandlerFunc) Handle(ctx context.Context, msg interface{}) error {
	return f(ctx, msg)
}

type DispatchCategory int

const (
	DispatchCategoryHandleParrllel DispatchCategory = iota
	DispatchCategoryHandleSerial
)

// Dispatch is a middleman between the Reader and Processor.
type Dispatch struct {
	reader            Reader
	handler           Handler
	isHandlerParallel bool
	wg                WaitGroup
	ctx               context.Context
}

func NewDispatch(reader Reader, handler Handler) *Dispatch {
	return NewDispatch3(reader, handler, DispatchCategoryHandleParrllel)
}
func NewDispatch3(reader Reader, handler Handler, category DispatchCategory) *Dispatch {
	var isHandlerParallel bool
	if category == DispatchCategoryHandleParrllel {
		isHandlerParallel = true
	}

	return &Dispatch{
		reader:            reader,
		handler:           handler,
		isHandlerParallel: isHandlerParallel,
	}
}
func (d *Dispatch) Context() context.Context {
	if d.ctx != nil {
		return d.ctx
	}
	return context.Background()
}
func (d *Dispatch) WithContext(ctx context.Context) *Dispatch {
	if ctx == nil {
		panic("nil context")
	}
	d2 := new(Dispatch)
	*d2 = *d
	d2.ctx = ctx

	return d2
}
func (d *Dispatch) done() bool {
	select {
	case <-d.Context().Done():
		return true
	default:
		return false
	}
}
func (d *Dispatch) Read() (interface{}, error) {
	return d.reader.Read(d.Context())
}
func (d *Dispatch) Handle(msg interface{}) error {
	fn := func(wg WaitGroup) error {
		wg.Add(1)
		defer wg.Done()
		return d.handler.Handle(d.Context(), msg)
	}
	if d.isHandlerParallel {
		go fn((d.waitGroup()))
	}
	return fn(d.waitGroup())
}

// 遍历读取消息，并进行分发处理
func (d *Dispatch) Start() *Dispatch {
	func(wg WaitGroup) {
		wg.Add(1)
		defer wg.Done()
		for {
			msg, err := d.Read()
			if err != nil {
				break
			}
			// just dispatch non-nil msg
			if msg != nil {
				d.Handle(msg)
			}

			// break if dispatcher is canceled or done
			if d.done() {
				break
			}
		}

	}(d.waitGroup())
	return d
}

// make Dispatch joinable
// Join() blocks until all recv and handle workflows started after Join() is finished
// RECOMMECD : call Joinable() before Start() to join all workflows
func (d *Dispatch) Joinable() *Dispatch {
	if d.wg == nil {
		wg := &sync.WaitGroup{}
		wg.Add(1)
		d.wg = wg
	}
	return d
}

// make Dispatch unjoinable, as Join() return immediately when called
func (d *Dispatch) UnJoinable() *Dispatch {
	d.waitGroup().Done()
	if d.wg != nil {
		d.wg = nil
	}
	return d
}
func (d *Dispatch) waitGroup() WaitGroup {
	if d.wg != nil {
		return d.wg
	}
	return nullWG
}

// wait until all recv and handle workflows finished, such as join in Thread
func (d *Dispatch) Join() *Dispatch {
	func(wg WaitGroup) {
		wg.Done()
		wg.Wait()
	}(d.waitGroup())
	return d
}
