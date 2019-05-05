package resilience

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"os/exec"
)

type CommandSharedPtr struct {
	*SharedPtr
}

func NewCommandSharedPtr(ctx context.Context, cmd *exec.Cmd, l logrus.FieldLogger) *CommandSharedPtr {
	resilienceSharedPtr := &CommandSharedPtr{
		SharedPtr: NewSharedPtr(ctx, func() (Ptr, error) {
			if cmd == nil {
				return nil, fmt.Errorf("resillence cmd: empty value")
			}
			return NewCommand(cmd), nil
		}, l),
	}
	resilienceSharedPtr.withWatch()
	resilienceSharedPtr.WithBackgroundTask()
	return resilienceSharedPtr
}

func (g *CommandSharedPtr) GetUntilReady() (*Command, error) {
	x, err := g.SharedPtr.GetUntilReady()
	if err != nil {
		return nil, err
	}
	ffmpeg, ok := x.Value().(*Command)
	if ok {
		return ffmpeg, nil
	}
	return nil, fmt.Errorf("unexpected type %T", x)
}
func (g *CommandSharedPtr) GetWithRetry() (*Command, error) {
	x, err := g.SharedPtr.GetWithRetry()
	if err != nil {
		return nil, err
	}
	cmd, ok := x.Value().(*Command)
	if ok {
		return cmd, nil
	}
	return nil, fmt.Errorf("unexpected type %T", x)
}
func (g *CommandSharedPtr) Get() (*Command, error) {
	x := g.SharedPtr.Get()
	if x == nil {
		return nil, nil
	}
	ffmpeg, ok := x.Value().(*Command)
	if ok {
		return ffmpeg, nil
	}
	return nil, fmt.Errorf("unexpected type %T", x)
}

func (g *CommandSharedPtr) withWatch() {
	done := make(chan struct{})
	// watch cmd's state
	eventC := g.Watch()
	go func() {
		select {
		case <-g.Context().Done():
			g.GetLogger().
				WithField("module", "cmd").
				Infof("cmd has been shutdown")
			eventC <- EventClose
			return
		case <-done:
			g.GetLogger().
				WithField("module", "cmd").
				Warnf("cmd has been expired")
			eventC <- EventExpired
		}
	}()
	go func() {
		defer close(done)
		cmd, err := g.GetUntilReady()
		if err != nil {
			return
		}
		if err := cmd.Wait(); err != nil {
			return
		}
		return
	}()
}

func (g *CommandSharedPtr) Handle() error {
	cmd, err := g.Get()
	if err != nil {
		return err
	}
	return cmd.Handle()
}
