package resilience

import (
	"fmt"
	"os/exec"
	"syscall"
)

type Command struct {
	*exec.Cmd
}

func NewCommand(cmd *exec.Cmd) *Command {
	return &Command{
		Cmd: cmd,
	}
}
func (r *Command) Value() interface{} {
	return r
}

func (r *Command) Ready() error {
	if r == nil || r.Cmd == nil {
		return fmt.Errorf("command: empty value")
	}

	if r.Cmd.ProcessState != nil && r.Cmd.ProcessState.Exited() {
		return fmt.Errorf("command: exited already %s", r.Cmd.ProcessState.String())
	}
	return nil
}

func (r *Command) Close() {
	if r == nil || r.Cmd == nil {
		return
	}
	proc := r.Cmd.Process
	if proc != nil {
		_ = proc.Signal(syscall.SIGTERM)
		proc.Wait()
		// proc.Kill()
		// no need to close attached log file.
		// see "Wait releases any resources associated with the cmd."
		// if closer, ok := cmd.Stdout.(io.Closer); ok {
		// 	closer.Close()
		// 	logger.Printf("process:%v Stdout closed.", proc)
		// }
	}
	r.Cmd = nil
}
func (r *Command) Handle() error {
	if r == nil || r.Cmd == nil {
		return fmt.Errorf("command: empty value")
	}
	defer func() {
		r.Cmd = nil
	}()
	return r.Start()
}
