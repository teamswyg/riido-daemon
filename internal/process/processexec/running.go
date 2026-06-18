package processexec

import (
	"context"
	"io"
	"os/exec"
	"sync"

	"github.com/teamswyg/riido-daemon/internal/process"
)

type execRunning struct {
	cmd       *exec.Cmd
	cancel    context.CancelFunc
	stdout    chan []byte
	stderr    chan []byte
	exited    chan process.ExitStatus
	stdin     io.WriteCloser
	stdinOnce sync.Once
	stdinMu   *sync.Mutex
	termOnce  sync.Once
	forceOnce sync.Once
	done      chan struct{}
}

func newExecRunning(
	cmd *exec.Cmd,
	cancel context.CancelFunc,
	stdin io.WriteCloser,
	stdinMu *sync.Mutex,
) *execRunning {
	return &execRunning{
		cmd:     cmd,
		cancel:  cancel,
		stdout:  make(chan []byte, process.DefaultStdoutBuffer),
		stderr:  make(chan []byte, process.DefaultStderrBuffer),
		exited:  make(chan process.ExitStatus, 1),
		stdin:   stdin,
		stdinMu: stdinMu,
		done:    make(chan struct{}),
	}
}

func (r *execRunning) Stdout() <-chan []byte             { return r.stdout }
func (r *execRunning) Stderr() <-chan []byte             { return r.stderr }
func (r *execRunning) Exited() <-chan process.ExitStatus { return r.exited }
