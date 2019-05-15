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

func NewCommandSharedPtr(ctx context.Context, cmd func() *exec.Cmd, l logrus.FieldLogger) *CommandSharedPtr {
	resilienceSharedPtr := &CommandSharedPtr{
		SharedPtr: NewSharedPtr(ctx, func() (Ptr, error) {
			if cmd == nil {
				return nil, fmt.Errorf("resillence cmd: empty value")
			}

			return NewCommand(cmd()), nil
		}, l),
	}
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

func (g *CommandSharedPtr) Run() error {
	cmd, err := g.Get()
	if err != nil {
		return err
	}
	return cmd.Run()
}
