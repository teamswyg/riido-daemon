// Package processexec is the os/exec implementation of the process port.
//
// The implementation spawns a single child process via exec.CommandContext
// (so context cancellation terminates the process tree on Unix), then
// fans out stdout / stderr / exit through bounded channels. stdin
// writes go through a dedicated channel-backed pipe so the session
// actor never blocks on a full kernel pipe.
//
// Concurrency: os/exec owns stdout/stderr copy goroutines for the stream
// writers, and this package owns the exit waiter. Each stream has a
// channel-owned writer; the public RunningProcess accessors only return those
// channels.
// The public stream contract stays channel-owned. The kill/stdin close paths
// use small synchronization guards because they cross goroutine boundaries
// owned by os/exec pipes.
package processexec

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/teamswyg/riido-daemon/internal/process"
)

// ChildObserver is notified when a spawned child starts and exits, so a caller
// can track live provider process groups for orphan reaping (D6). pid is the
// child PID, which equals its process-group id under Setpgid.
type ChildObserver interface {
	OnSpawn(pid int)
	OnExit(pid int)
}

// New returns a process.Process that spawns via os/exec.
func New() process.Process { return &execProcess{} }

// NewWithObserver is New plus child-lifecycle notifications. A nil observer
// behaves like New.
func NewWithObserver(obs ChildObserver) process.Process { return &execProcess{obs: obs} }

type execProcess struct{ obs ChildObserver }

func (e *execProcess) Start(ctx context.Context, cmd process.Command) (process.RunningProcess, error) {
	if cmd.Executable == "" {
		return nil, errors.New("processexec: empty Executable")
	}

	cmdCtx, cancel := context.WithCancel(ctx)
	c := exec.CommandContext(cmdCtx, cmd.Executable, cmd.Args...)
	c.Env = mergeEnv(cmd.Env)
	c.Dir = cmd.Dir
	configureCommand(c)

	stdinPipe, err := c.StdinPipe()
	if err != nil {
		cancel()
		return nil, err
	}

	r := &execRunning{
		cmd:     c,
		cancel:  cancel,
		obs:     e.obs,
		stdout:  make(chan []byte, process.DefaultStdoutBuffer),
		stderr:  make(chan []byte, process.DefaultStderrBuffer),
		exited:  make(chan process.ExitStatus, 1),
		stdin:   stdinPipe,
		stdinMu: &sync.Mutex{},
		done:    make(chan struct{}),
	}
	c.Stdout = streamWriter{out: r.stdout}
	c.Stderr = streamWriter{out: r.stderr}

	if err := c.Start(); err != nil {
		cancel()
		return nil, err
	}

	if c.Process != nil {
		r.pid = c.Process.Pid
	}
	if e.obs != nil && r.pid > 0 {
		e.obs.OnSpawn(r.pid)
	}

	go r.killOnContext(cmdCtx.Done())
	go r.waitExit()

	return r, nil
}

func mergeEnv(overrides []string) []string {
	if len(overrides) == 0 {
		return nil
	}
	env := os.Environ()
	indexByKey := make(map[string]int, len(env)+len(overrides))
	for i, entry := range env {
		key, _, ok := strings.Cut(entry, "=")
		if ok {
			indexByKey[key] = i
		}
	}
	for _, entry := range overrides {
		key, _, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		if i, exists := indexByKey[key]; exists {
			env[i] = entry
			continue
		}
		indexByKey[key] = len(env)
		env = append(env, entry)
	}
	return env
}

type execRunning struct {
	cmd       *exec.Cmd
	cancel    context.CancelFunc
	obs       ChildObserver
	pid       int
	stdout    chan []byte
	stderr    chan []byte
	exited    chan process.ExitStatus
	stdin     io.WriteCloser
	stdinOnce sync.Once
	stdinMu   *sync.Mutex
	killOnce  sync.Once
	done      chan struct{}
}

type streamWriter struct {
	out chan<- []byte
}

func (w streamWriter) Write(p []byte) (int, error) {
	chunk := make([]byte, len(p))
	copy(chunk, p)
	w.out <- chunk
	return len(p), nil
}

func (r *execRunning) Stdout() <-chan []byte             { return r.stdout }
func (r *execRunning) Stderr() <-chan []byte             { return r.stderr }
func (r *execRunning) Exited() <-chan process.ExitStatus { return r.exited }

func (r *execRunning) WriteStdin(b []byte) error {
	r.stdinMu.Lock()
	defer r.stdinMu.Unlock()
	if r.stdin == nil {
		return errors.New("processexec: stdin closed")
	}
	_, err := r.stdin.Write(b)
	return err
}

func (r *execRunning) CloseStdin() error {
	var err error
	r.stdinOnce.Do(func() {
		r.stdinMu.Lock()
		defer r.stdinMu.Unlock()
		if r.stdin != nil {
			err = r.stdin.Close()
			r.stdin = nil
		}
	})
	return err
}

func (r *execRunning) Kill(_ context.Context) error {
	r.cancel()
	r.terminateProcessGroup()
	return nil
}

func (r *execRunning) killOnContext(ctxDone <-chan struct{}) {
	select {
	case <-ctxDone:
		r.terminateProcessGroup()
	case <-r.done:
	}
}

func (r *execRunning) terminateProcessGroup() {
	r.killOnce.Do(func() {
		terminateCommand(r.cmd)
	})
}

func (r *execRunning) waitExit() {
	err := r.cmd.Wait()
	if r.obs != nil && r.pid > 0 {
		// The process (and its group, after terminateCommand) is reaped; drop it
		// from the live registry so only orphans from an unclean exit remain.
		r.obs.OnExit(r.pid)
	}
	close(r.done)
	close(r.stdout)
	close(r.stderr)
	code := r.cmd.ProcessState.ExitCode()
	if code < 0 && err != nil {
		// Negative exit code = killed by signal; preserve err.
		code = 137 // conventional kill-via-SIGKILL exit code
	}
	r.exited <- process.ExitStatus{Code: code, Err: err}
	close(r.exited)
}
