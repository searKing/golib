package resilience

import (
	"context"
	"fmt"
	"github.com/searKing/golib/x/log"
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
	ErrEmptyValue      = fmt.Errorf("empty value")
	ErrAlreadyShutdown = fmt.Errorf("already shutdown")
	ErrNotReady        = fmt.Errorf("not ready")
)

type Ptr interface {
	Value() interface{} //actual instance
	Ready() error
	Close()
}

type SharedPtr struct {
	// New optionally specifies a function to generate
	// a value when Get would otherwise return nil.
	// It may not be changed concurrently with calls to Get.
	New func() (Ptr, error)
	*log.FieldLogger

	TaskMaxDuration time.Duration
	Timeout         time.Duration
	// ctx is either the client or server context. It should only
	// be modified via copying the whole Request using WithContext.
	// It is unexported to prevent people from using Context wrong
	// and mutating the contexts held by callers of the same request.
	ctx context.Context

	x      Ptr
	taskC  chan *Task
	tasks  map[*Task]struct{}
	eventC chan Event

	mu sync.Mutex
}

func NewSharedPtr(ctx context.Context, new func() (Ptr, error), l logrus.FieldLogger) *SharedPtr {
	return &SharedPtr{
		New:             new,
		FieldLogger:     log.New(l),
		TaskMaxDuration: DefaultResilienceTaskMaxDuration,
		Timeout:         DefaultResilienceTimeout,
		ctx:             ctx,
	}
}

func NewSharedPtrFunc(ctx context.Context,
	new func() (interface{}, error),
	ready func(x interface{}) error,
	close func(x interface{}), l logrus.FieldLogger) *SharedPtr {
	return NewSharedPtr(ctx, WithFuncNewer(new, ready, close), l)
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

type Event int

const (
	EventNew     Event = iota // new and start
	EventClose                // close
	EventExpired              // restart
)

func (g *SharedPtr) InShutdown() bool {
	select {
	case <-g.Context().Done():
		return true
	default:
		return false
	}
}

func (g *SharedPtr) getTaskC() chan *Task {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.taskC == nil {
		g.taskC = make(chan *Task)
	}
	return g.taskC
}

func (g *SharedPtr) AddTask(task *Task) {
	if task == nil || task.Handle == nil {
		return
	}
	if g.InShutdown() {
		return
	}
	task.ctx = g.Context()
	g.getTaskC() <- task
}
func (g *SharedPtr) AddTaskFuncAsConstruct(handle func() error) {
	if handle == nil || g == nil {
		return
	}
	if g.InShutdown() {
		return
	}
	g.getTaskC() <- &Task{
		Type:   TaskTypeConstruct,
		Handle: handle,
		ctx:    g.Context(),
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
			case task, ok := <-g.getTaskC():
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
							select {
							case <-task.Context().Done():
								g.mu.Lock()
								defer g.mu.Unlock()
								delete(g.tasks, task)
							}
							switch task.State {
							case TaskStateNew:
								go func() {
									<-time.After(g.TaskMaxDuration)
									g.getTaskC() <- task
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
						g.GetLogger().WithError(err).Warn("Retry failed...")
						continue
					}
					if event == EventExpired {
						g.recoveryTask(false)
					}
					g.GetLogger().Infof("Retry success...")
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

	if g.InShutdown() {
		return ErrAlreadyShutdown
	}
	x := g.Get()
	if x != nil {
		return x.Ready()
	}
	return ErrEmptyValue
}

// std::shared_ptr.get() until ptr is ready & std::shared_ptr.make_unique() if necessary
func (g *SharedPtr) GetUntilReady() (Ptr, error) {
	err := Retry(g.Context(), g.GetLogger(), g.TaskMaxDuration, g.Timeout, func() error {
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
	err := Retry(g.Context(), g.GetLogger(), g.TaskMaxDuration, g.Timeout, func() error {
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
				g.getTaskC() <- task
			}
		}
	}()
}