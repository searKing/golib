package resilience

import (
	"context"
	"fmt"
	"github.com/searKing/golib/sync_/atomic_"
	"github.com/searKing/golib/x/log"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

const (
	// DefaultConnectTimeout is the default timeout to establish a connection to
	// a ZooKeeper node.
	DefaultResilienceConstructTimeout = 0
	// DefaultSessionTimeout is the default timeout to keep the current
	// ZooKeeper session alive during a temporary disconnect.
	DefaultResilienceTaskMaxRetryDuration = 15 * time.Second

	DefaultTaskRetryTimeout      = 1 * time.Second
	DefaultTaskRescheduleTimeout = 1 * time.Second
)

var (
	ErrEmptyValue      = fmt.Errorf("empty value")
	ErrAlreadyShutdown = fmt.Errorf("already shutdown")
	ErrNotReady        = fmt.Errorf("not ready")
)

type SharedPtr struct {
	// New optionally specifies a function to generate
	// a value when Get would otherwise return nil.
	// It may not be changed concurrently with calls to Get.
	New func() (Ptr, error)
	*log.FieldLogger

	// to judge whether Get&Construct is timeout
	ConstructTimeout time.Duration
	// MaxDuration for retry if tasks failed
	TaskMaxRetryDuration time.Duration

	// ctx is either the client or server context. It should only
	// be modified via copying the whole Request using WithContext.
	// It is unexported to prevent people from using Context wrong
	// and mutating the contexts held by callers of the same request.
	ctx context.Context

	x      Ptr
	taskC  chan *Task
	tasks  map[*Task]struct{}
	eventC chan Event

	backgroundStopped atomic_.Bool

	mu sync.Mutex
}

func NewSharedPtr(ctx context.Context, new func() (Ptr, error), l logrus.FieldLogger) *SharedPtr {
	return &SharedPtr{
		New:                  new,
		FieldLogger:          log.New(l),
		TaskMaxRetryDuration: DefaultResilienceTaskMaxRetryDuration,
		ConstructTimeout:     DefaultResilienceConstructTimeout,
		ctx:                  ctx,
	}
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
	if g == nil || task == nil || task.Handle == nil {
		return
	}
	if g.InShutdown() {
		return
	}
	task.ctx = g.Context()

	go func() {
		if task.Type.Construct {
			_, err := g.GetWithRetry()
			if err != nil {
				return
			}
		} else {
			_, err := g.GetUntilReady()
			if err != nil {
				return
			}
		}
		g.getTaskC() <- task
	}()
}
func (g *SharedPtr) AddTaskFunc(taskType TaskType, handle func() error, descriptions ...string) {
	if handle == nil || g == nil {
		return
	}
	g.AddTask(&Task{
		Description:    strings.Join(descriptions, ""),
		Type:           taskType,
		Handle:         handle,
		RetryDuration:  DefaultTaskRetryTimeout,
		RepeatDuration: DefaultTaskRescheduleTimeout,
		ctx:            g.Context(),
	})
}
func (g *SharedPtr) AddTaskFuncAsConstruct(handle func() error, descriptions ...string) {
	if handle == nil || g == nil {
		return
	}
	g.AddTaskFunc(TaskType{Construct: true}, handle, descriptions...)
}
func (g *SharedPtr) AddTaskFuncAsConstructRepeat(handle func() error, descriptions ...string) {
	if handle == nil || g == nil {
		return
	}
	g.AddTaskFunc(TaskType{Construct: true, Repeat: true}, handle, descriptions...)
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
					g.GetLogger().WithField("event", event.String()).Infof("handle event success...")
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
	err := Retry(g.Context(), g.GetLogger(), g.TaskMaxRetryDuration, g.ConstructTimeout, func() error {
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
	err := Retry(g.Context(), g.GetLogger(), g.TaskMaxRetryDuration, g.ConstructTimeout, func() error {
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
		g.backgroundTask(true)
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

			if !task.Type.Drop {
				task.State = TaskStateDormancy
			}

			if task.State == TaskStateDormancy {
				task.State = TaskStateNew
				g.getTaskC() <- task
			}
		}
	}()
}

func (g *SharedPtr) backgroundTask(locked bool) {
	swapped := g.backgroundStopped.CAS(false, true)
	if !swapped {
		return
	}

	defer func() {
		if !locked {
			g.mu.Lock()
			defer g.mu.Unlock()
		}
		if g.tasks == nil {
			g.tasks = make(map[*Task]struct{})
		}
	}()
	go func() {
		defer func() {
			g.backgroundStopped.Store(false)
		}()
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
				if task.State == TaskStateRunning {
					g.GetLogger().WithField("task", task).
						Warn("task is running already, ignore duplicate schedule...")
					continue
				}
				// verify whether task is duplicated
				// store task
				addedTask := func() (added bool) {
					g.mu.Lock()
					defer g.mu.Unlock()
					if g.tasks == nil {
						g.tasks = make(map[*Task]struct{})
					}
					_, has := g.tasks[task]
					return has
				}()

				addTask := func() {
					g.mu.Lock()
					defer g.mu.Unlock()
					if g.tasks == nil {
						g.tasks = make(map[*Task]struct{})
					}
					g.tasks[task] = struct{}{}
				}

				deleteTask := func() {
					g.mu.Lock()
					defer g.mu.Unlock()
					delete(g.tasks, task)
				}
				if addedTask {
					g.GetLogger().WithField("task", task).
						Warn("task is added already, ignore duplicate schedule...")
					continue
				}
				if task.Type.Construct {
					if _, err := g.GetWithRetry(); err != nil {
						g.GetLogger().WithField("task", task).
							Warn("task is added but not scheduled, new has not been called yet...")
						continue
					}
				} else {
					if _, err := g.GetUntilReady(); err != nil {
						g.GetLogger().WithField("task", task).
							Warn("task is added but not scheduled, not ready yet...")
						continue
					}
				}

				// Handle task
				go func() {
					if task.State == TaskStateNew {
						addTask()
						task.State = TaskStateRunning
						g.GetLogger().WithField("task", task).Info("task is running now...")

						// execute the task and refresh the state
						func() {
							if task.Handle == nil {
								task.State = TaskStateDoneNormally
								return
							}
							if err := task.Handle(); err != nil {
								task.State = TaskStateDoneErrorHappened
								g.GetLogger().WithField("task", task).WithError(err).
									Warnf("task is done failed...")
								return
							}
							g.GetLogger().WithField("task", task).
								Info("task is done successfully...")
							task.State = TaskStateDoneNormally
						}()

						// handle completed execution and refresh the state
						func() {
							waitBeforeRepeat := func() {
								g.GetLogger().WithField("task", task).
									Warnf("Reschedule to repeat in %s...", task.RepeatDuration)
								<-time.After(task.RepeatDuration)
							}
							waitBeforeRecover := func() {
								g.GetLogger().WithField("task", task).
									Warnf("Reschedule to recover in %s...", task.RetryDuration)
								<-time.After(task.RetryDuration)
							}
							select {
							case <-task.Context().Done():
								task.State = TaskStateDeath
								return
							default:

								// Drop
								if task.Type.Drop && !task.Type.Retry {
									task.State = TaskStateDeath
									return
								}

								if task.Type.Drop && task.Type.Retry {
									if task.State == TaskStateDoneErrorHappened {
										waitBeforeRecover()
										task.State = TaskStateNew
										return
									}
									task.State = TaskStateDeath
									return
								}

								// Repeat
								if task.Type.Repeat && !task.Type.Construct {
									waitBeforeRepeat()
									task.State = TaskStateNew
									return
								}

								if task.Type.Repeat && task.Type.Construct {
									g.GetLogger().WithField("task", task).
										Warnf("task is delete and restart all tasks...")
									deleteTask() // don't recover this task, this task will be added later
									g.Reset()
									go func() {
										_, _ = g.GetWithRetry()
									}()
									waitBeforeRepeat()
									task.State = TaskStateNew
									return
								}

								// Construct && !Drop && !Repeat
								if task.Type.Construct {
									//Retry
									if task.Type.Retry && task.State == TaskStateDoneErrorHappened {
										g.GetLogger().WithField("task", task).
											Warnf("task is delete and restart all tasks...")
										deleteTask() // don't recover this task, this task will be added later
										g.Reset()
										go func() {
											_, _ = g.GetWithRetry()
										}()
										waitBeforeRecover()
										task.State = TaskStateNew
										return
									}
									task.State = TaskStateDormancy
									return
								}
								task.State = TaskStateDeath
								return
							}
						}()
						// complete the task's life cycle
						func() {
							select {
							case <-task.Context().Done():
								g.GetLogger().WithField("task", task).
									Info("Go to death now...")
								go deleteTask()
								return
							default:
								switch task.State {
								case TaskStateNew:
									go func() {
										deleteTask()
										g.GetLogger().WithField("task", task).
											Infof("Reschedule now...")
										g.getTaskC() <- task
									}()
								case TaskStateDormancy:
									g.GetLogger().WithField("task", task).
										Infof("task is done,  go to dormancy...")
								case TaskStateDeath:
									g.GetLogger().WithField("task", task).
										Infof("task is dead,  go to death...")
									fallthrough
								default:
									g.GetLogger().WithField("task", task).
										Info("Go to death now...")
									go deleteTask()
								}
							}
						}()
					}
				}()
			}
		}

	}()
}
