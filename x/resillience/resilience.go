package resilience

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

const (
	// DefaultConnectTimeout is the default timeout to establish a connection to
	// a ZooKeeper node.
	DefaultResilienceTimeout = 0
	// DefaultSessionTimeout is the default timeout to keep the current
	// ZooKeeper session alive during a temporary disconnect.
	DefaultResilienceTaskMaxDuration = 15 * time.Second
)

var (
	ErrEmptyValue = fmt.Errorf("empty value")
	ErrNotReady   = fmt.Errorf("not ready")
)

type Ptr interface {
	Value() interface{} //actual instance
	Ready() error
	Close()
}
type emptyResilience int

func (r *emptyResilience) Value() interface{} {
	return nil
}

func (r *emptyResilience) Ready() error {
	return nil
}

func (r *emptyResilience) Close() {
	return
}

var (
	background = new(emptyResilience)
	todo       = new(emptyResilience)
)

// Background returns a non-nil, empty Context. It is never canceled, has no
// values, and has no deadline. It is typically used by the main function,
// initialization, and tests, and as the top-level Context for incoming
// requests.
func Background() Ptr {
	return background
}

// TODO returns a non-nil, empty Context. Code should use context.TODO when
// it's unclear which Context to use or it is not yet available (because the
// surrounding function has not yet been extended to accept a Context
// parameter).
func TODO() Ptr {
	return todo
}

type funcResilience struct {
	x     interface{}
	ready func(x interface{}) error
	close func(x interface{})
}

func (r *funcResilience) Value() interface{} {
	if r == nil {
		return nil
	}
	return r.x
}

func (r *funcResilience) Ready() error {
	if r == nil {
		return nil
	}
	if r.ready == nil {
		return nil
	}
	return r.ready(r.x)
}

func (r *funcResilience) Close() {
	if r == nil {
		return
	}
	if r.close == nil {
		return
	}
	r.close(r.x)
}

func WithFunc(x interface{}, ready func(x interface{}) error,
	close func(x interface{})) (Ptr, error) {
	return &funcResilience{
		x:     x,
		ready: ready,
		close: close,
	}, nil
}

func WithFuncNewer(new func() (interface{}, error),
	ready func(x interface{}) error,
	close func(x interface{})) func() (Ptr, error) {
	return func() (Ptr, error) {
		if new == nil {
			return nil, ErrEmptyValue
		}
		x, err := new()
		if err != nil {
			return nil, err
		}
		return &funcResilience{
			x:     x,
			ready: ready,
			close: close,
		}, nil
	}
}

type TaskType int

const (
	TaskTypeDisposable      TaskType = iota // Task will be executed once and dropped whether it's successful or not
	TaskTypeDisposableRetry                 // Task will be executed and dropped until it's successful
	TaskTypeRepeat                          // Task will be executed again and again
	TaskTypeConstruct                       // Task will be executed once after New is called
	TaskTypeButt
)

type TaskState int

const (
	TaskStateNew               TaskState = iota // Task state for a task which has not yet started.
	TaskStateRunning                            // Task state for a running task. A task in the running state is executing in the Go routine but it may be waiting for other resources from the operating system such as processor.
	TaskStateDoneErrorHappened                  // Task state for a terminated state. The task has completed execution with some errors happened
	TaskStateDoneNormally                       // Task state for a terminated state. The task has completed execution normally
	TaskStateDormancy                           // Task state for a terminated state. The task has completed execution normally and will be started if New's called
	TaskStateDeath                              // Task state for a terminated state. The task has completed execution normally and will be started if New's called
	TaskStateButt
)

type Task struct {
	Type   TaskType
	State  TaskState
	Handle func() error

	ctx context.Context
}

//
// The returned context is always non-nil; it defaults to the
// background context.
func (g *Task) Context() context.Context {
	if g.ctx != nil {
		return g.ctx
	}
	return context.Background()
}

// WithContext returns a shallow copy of r with its context changed
// to ctx. The provided ctx must be non-nil.
func (g *Task) WithContext(ctx context.Context) *Task {
	if ctx == nil {
		panic("nil context")
	}
	r2 := new(Task)
	*r2 = *g
	r2.ctx = ctx
	return r2
}

type SharedPtr struct {
	// New optionally specifies a function to generate
	// a value when Get would otherwise return nil.
	// It may not be changed concurrently with calls to Get.
	New func() (Ptr, error)

	TaskMaxDuration time.Duration
	Timeout         time.Duration
	// ctx is either the client or server context. It should only
	// be modified via copying the whole Request using WithContext.
	// It is unexported to prevent people from using Context wrong
	// and mutating the contexts held by callers of the same request.
	ctx context.Context

	L logrus.FieldLogger

	x      Ptr
	taskC  chan *Task
	tasks  map[*Task]struct{}
	eventC chan Event

	mu sync.Mutex
}

func NewSharedPtr(new func() (Ptr, error), l logrus.FieldLogger) *SharedPtr {
	return &SharedPtr{
		New:             new,
		L:               l,
		TaskMaxDuration: DefaultResilienceTaskMaxDuration,
		Timeout:         DefaultResilienceTimeout,
	}
}
func NewSharedPtrFunc(
	new func() (interface{}, error),
	ready func(x interface{}) error,
	close func(x interface{}), l logrus.FieldLogger) *SharedPtr {
	return NewSharedPtr(WithFuncNewer(new, ready, close), l)
}

//
// The returned context is always non-nil; it defaults to the
// background context.
func (g *SharedPtr) Context() context.Context {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.ctx != nil {
		return g.ctx
	}
	return context.Background()
}

// WithContext returns a shallow copy of r with its context changed
// to ctx. The provided ctx must be non-nil.
func (g *SharedPtr) WithContext(ctx context.Context) {
	if ctx == nil {
		panic("nil context")
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.ctx = ctx
	return
}

type Event int

const (
	EventNew Event = iota
	EventClose
	EventExpired
)

func (g *SharedPtr) GetTaskC() chan *Task {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.taskC == nil {
		g.taskC = make(chan *Task)
	}
	return g.taskC
}

func (g *SharedPtr) AddTask(task *Task) {
	if task == nil {
		return
	}
	g.GetTaskC() <- task
}
func (g *SharedPtr) AddTaskFuncAsConstruct(ctx context.Context, handle func() error) {
	if handle == nil {
		return
	}
	g.GetTaskC() <- &Task{
		Type:   TaskTypeConstruct,
		Handle: handle,
		ctx:    ctx,
	}
}

func (g *SharedPtr) WithBackgroundTask() {
	defer func() {
		g.mu.Lock()
		defer g.mu.Unlock()
		g.tasks = make(map[*Task]struct{})
	}()
	go func() {
	L:
		for {
			select {
			case <-g.Context().Done():
				break L
			case task, ok := <-g.GetTaskC():
				if !ok {
					break L
				}
				if task == nil {
					continue
				}
				func() {
					g.mu.Lock()
					defer g.mu.Unlock()
					if g.tasks == nil {
						g.tasks = make(map[*Task]struct{})
					}
					g.tasks[task] = struct{}{}
				}()
				go func() {
					if task.State == TaskStateNew {
						task.State = TaskStateRunning

						// execute the task and refresh the state
						func() {
							if task.Handle == nil {
								task.State = TaskStateDoneNormally
								return
							}
							if err := task.Handle(); err != nil {
								task.State = TaskStateDoneErrorHappened
								return
							}
							task.State = TaskStateDoneNormally
						}()

						// handle completed execution and refresh the state
						func() {
							select {
							case <-task.Context().Done():
								task.State = TaskStateDeath
							default:
								switch task.Type {
								case TaskTypeDisposable:
									task.State = TaskStateDeath
								case TaskTypeDisposableRetry:
									if task.State == TaskStateDoneErrorHappened {
										task.State = TaskStateNew
									} else {
										task.State = TaskStateDeath
									}
								case TaskTypeRepeat:
									task.State = TaskStateNew
								case TaskTypeConstruct:
									task.State = TaskStateDormancy
								default:
									task.State = TaskStateDeath
								}
							}
						}()

						// complete the task's life cycle
						func() {
							switch task.State {
							case TaskStateNew:
								go func() {
									<-time.After(g.TaskMaxDuration)
									g.GetTaskC() <- task
								}()
							case TaskStateDormancy:
							case TaskStateDeath:
								fallthrough
							default:
								g.mu.Lock()
								defer g.mu.Unlock()
								delete(g.tasks, task)
							}
						}()
					}
				}()
			}
		}

	}()
}

func (g *SharedPtr) event() chan Event {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.eventC == nil {
		g.eventC = make(chan Event)
	}
	return g.eventC
}

func (g *SharedPtr) Watch() chan<- Event {
	eventC := g.event()
	go func() {
	L:
		for {
			select {
			case <-g.Context().Done():
				break L
			case event, ok := <-eventC:
				if !ok {
					break L
				}
				switch event {
				case EventNew, EventExpired:
					if event == EventExpired {
						g.Reset()
					}
					// New x
					_, err := g.GetWithRetry()
					if err != nil {
						g.L.WithError(err).Warn("Retry failed...")
						continue
					}
					if event == EventExpired {
						g.recoveryTask(false)
					}
					g.L.Infof("Retry success...")
				case EventClose:
					g.Reset()
				}
			}
		}
	}()
	return eventC
}

func (g *SharedPtr) Ready() error {
	if g == nil {
		return ErrEmptyValue
	}
	x := g.Get()
	if x != nil {
		return x.Ready()
	}
	return ErrEmptyValue
}

// std::shared_ptr.get() until ptr is ready & std::shared_ptr.make_unique() if necessary
func (g *SharedPtr) GetUntilReady() (Ptr, error) {
	err := Retry(g.Context(), g.L, g.TaskMaxDuration, g.Timeout, func() error {
		x := g.Get()
		if x != nil {
			// check  if x is ready
			if err := x.Ready(); err != nil {
				// until ready
				return err
			}
			return nil
		}

		// New x
		x, err := g.allocate()
		if err != nil {
			return err
		}
		if x == nil {
			return ErrEmptyValue
		}
		return ErrNotReady
	})
	return g.Get(), err
}

// std::shared_ptr.get() & std::shared_ptr.make_unique() if necessary
func (g *SharedPtr) GetWithRetry() (Ptr, error) {
	// if allocated, return now
	if x := g.Get(); x != nil {
		return x, nil
	}

	// New x
	err := Retry(g.Context(), g.L, g.TaskMaxDuration, g.Timeout, func() error {
		_, err := g.allocate()
		return err
	})
	if err != nil {
		return nil, err
	}

	return g.Get(), nil
}

// std::shared_ptr.release()
func (g *SharedPtr) Release() Ptr {
	g.mu.Lock()
	defer g.mu.Unlock()
	x := g.x
	g.x = nil
	return x
}

// std::shared_ptr.reset()
func (g *SharedPtr) Reset() {
	x := g.Release()
	if x != nil {
		x.Close()
	}
	return
}

// std::shared_ptr.get()
func (g *SharedPtr) Get() Ptr {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.x
}

func (g *SharedPtr) allocate() (Ptr, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.allocateLocked()
}

func (g *SharedPtr) allocateLocked() (Ptr, error) {
	if g.x != nil {
		return g.x, nil
	}
	if g.New != nil {
		x, err := g.New()
		if err != nil {
			return g.x, err
		}
		g.x = x
		g.recoveryTask(true)
	}
	return g.x, nil

}

func (g *SharedPtr) recoveryTask(locked bool) {
	tasks := func() map[*Task]struct{} {
		if !locked {
			g.mu.Lock()
			defer g.mu.Unlock()
		}
		tasks := g.tasks
		g.tasks = nil
		return tasks
	}()
	go func() {
	L:
		for task := range tasks {
			if task == nil {
				continue
			}
			select {
			case <-g.Context().Done():
				break L
			case <-task.Context().Done():
				break L
			default:
			}
			if task.Type == TaskTypeConstruct {
				task.State = TaskStateDormancy
			}

			if task.State == TaskStateDormancy {
				task.State = TaskStateNew
				g.GetTaskC() <- task
			}
		}
	}()
}
