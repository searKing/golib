package exec

import (
	"context"
	"io"
	"os"
	"os/exec"
)

type commandSercer struct {
	cmd    *exec.Cmd
	handle func(reader io.Reader)
	ctx    context.Context
	done   context.CancelFunc
}

func newCommandServer(parent context.Context, stop context.CancelFunc, handle func(reader io.Reader), name string, args ...string) (*commandSercer, error) {
	if parent == nil {
		parent = context.Background()
	}
	if stop == nil {
		stop = func() {}
	}
	if handle == nil {
		handle = func(reader io.Reader) {}
	}

	cs := &commandSercer{
		cmd:    exec.Command(name, args...),
		handle: handle,
		ctx:    parent,
		done:   stop,
	}

	r, err := cs.cmd.StdoutPipe()
	if err != nil {
		cs.Stop()
		return nil, err
	}
	go cs.watch(r)
	return cs, nil
}

func (cs *commandSercer) wait() error {
	select {
	case <-cs.ctx.Done():
		return cs.ctx.Err()
	}
	return nil
}

func (cs *commandSercer) watch(r io.Reader) {
	cs.handle(r)
	cs.cmd.Wait()
	cs.done()
}
func (cs *commandSercer) Stop() {
	cs.cmd.Process.Signal(os.Interrupt)
	cs.done()
}
