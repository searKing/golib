package resilience

import (
	"context"
	"fmt"
	"time"
)

type TaskType int

const (
	TaskTypeDisposable      TaskType = iota // Task will be executed once and dropped whether it's successful or not
	TaskTypeDisposableRetry                 // Task will be executed and dropped until it's successful
	TaskTypeRepeat                          // Task will be executed again and again, even New is called
	TaskTypeConstruct                       // Task will be executed once after New is called, don't wait for ready
	TaskTypeConstructRepeat                 // Task will be executed again and again after New is called, don't wait for ready
	TaskTypeButt
)

func (t TaskType) String() string {
	return taskType[t]
}

var taskType = map[TaskType]string{
	TaskTypeDisposable:      "disposable",
	TaskTypeDisposableRetry: "disposable_retry",
	TaskTypeRepeat:          "repeat",
	TaskTypeConstruct:       "construct",
	TaskTypeConstructRepeat: "construct_repeat",
}

type TaskState int

const (
	TaskStateNew               TaskState = iota // Task state for a task which has not yet started.
	TaskStateRunning                            // Task state for a running task. A task in the running state is executing in the Go routine but it may be waiting for other resources from the operating system such as processor.
	TaskStateDoneErrorHappened                  // Task state for a terminated state. The task has completed execution with some errors happened
	TaskStateDoneNormally                       // Task state for a terminated state. The task has completed execution normally
	TaskStateDormancy                           // Task state for a terminated state. The task has completed execution normally and will be started if New's called
	TaskStateDeath                              // Task state for a terminated state. The task has completed execution normally and will never be started again
	TaskStateButt
)

func (t TaskState) String() string {
	return taskState[t]
}

var taskState = map[TaskState]string{
	TaskStateNew:               "new",
	TaskStateRunning:           "running",
	TaskStateDoneErrorHappened: "done_error_happened",
	TaskStateDoneNormally:      "done_normally",
	TaskStateDormancy:          "dormancy",
	TaskStateDeath:             "death",
}

type Task struct {
	Type        TaskType
	State       TaskState
	Description string // for debug
	Handle      func() error

	RepeatDuration time.Duration
	RetryDuration  time.Duration

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

func (g *Task) String() string {
	if g == nil {
		return "empty task"
	}
	return fmt.Sprintf("%s-%s-%s", g.Type, g.State, g.Description)
}
