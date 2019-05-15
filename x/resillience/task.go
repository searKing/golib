package resilience

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type TaskType struct {
	Drop      bool // task will be dropped
	Retry     bool // task will be retried if error happens
	Construct bool // task will be called after New
	Repeat    bool // Task will be executed again and again
}

func (t TaskType) String() string {
	var b strings.Builder
	if t.Drop {
		b.WriteString("drop")
	}
	if t.Retry {
		if b.String() != "" {
			b.WriteRune('-')
		}
		b.WriteString("retry")
	}
	if t.Construct {
		if b.String() != "" {
			b.WriteRune('-')
		}
		b.WriteString("construct")
	}
	if t.Repeat {
		if b.String() != "" {
			b.WriteRune('-')
		}
		b.WriteString("repeat")
	}
	return b.String()
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
