package resilience

import "context"

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

	ctx        context.Context
	inShutdown bool
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
