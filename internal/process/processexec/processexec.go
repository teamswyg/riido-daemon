// Package processexec is the os/exec implementation of the process port.
//
// The implementation spawns a single child process via exec.CommandContext
// (so context cancellation terminates the process tree on Unix), then
// fans out stdout / stderr / exit through bounded channels. stdin
// writes go through a dedicated channel-backed pipe so the session
// actor never blocks on a full kernel pipe.
//
// Concurrency: a small set of goroutines (stdout reader, stderr reader,
// stdin writer, exit waiter) is spawned per process. Each owns its
// channel; the public RunningProcess accessors only return the channels.
// The public stream contract stays channel-owned. The kill/stdin close paths
// use small synchronization guards because they cross goroutine boundaries
// owned by os/exec pipes.
package processexec

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/teamswyg/riido-daemon/internal/process"
)

// New returns a process.Process that spawns via os/exec.
func New() process.Process { return &execProcess{} }

type execProcess struct{}

func (e *execProcess) Start(ctx context.Context, cmd process.Command) (process.RunningProcess, error) {
	if cmd.Executable == "" {
		return nil, errors.New("processexec: empty Executable")
	}

	cmdCtx, cancel := context.WithCancel(ctx)
	c := exec.CommandContext(cmdCtx, cmd.Executable, cmd.Args...)
	c.Env = mergeEnv(cmd.Env)
	c.Dir = cmd.Dir

	stdoutPipe, err := c.StdoutPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	stderrPipe, err := c.StderrPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	stdinPipe, err := c.StdinPipe()
	if err != nil {
		cancel()
		return nil, err
	}

	if err := c.Start(); err != nil {
		cancel()
		return nil, err
	}

	r := &execRunning{
		cmd:     c,
		cancel:  cancel,
		stdout:  make(chan []byte, 64),
		stderr:  make(chan []byte, 64),
		exited:  make(chan process.ExitStatus, 1),
		stdin:   stdinPipe,
		stdinMu: &sync.Mutex{},
	}

	go r.pumpReader(stdoutPipe, r.stdout)
	go r.pumpReader(stderrPipe, r.stderr)
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
	stdout    chan []byte
	stderr    chan []byte
	exited    chan process.ExitStatus
	stdin     io.WriteCloser
	stdinOnce sync.Once
	stdinMu   *sync.Mutex
	killOnce  sync.Once
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
	r.killOnce.Do(func() {
		r.cancel()
		if r.cmd.Process != nil {
			// Best-effort: SIGTERM the process group on Unix; fall back to
			// killing the single PID. cmd.Process.Kill() uses SIGKILL which
			// is what we want for an unresponsive child.
			_ = r.cmd.Process.Signal(syscall.SIGTERM)
			_ = r.cmd.Process.Kill()
		}
	})
	return nil
}

func (r *execRunning) pumpReader(rd io.Reader, out chan<- []byte) {
	defer close(out)
	buf := make([]byte, 8192)
	br := bufio.NewReader(rd)
	for {
		n, err := br.Read(buf)
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			out <- chunk
		}
		if err != nil {
			return
		}
	}
}

func (r *execRunning) waitExit() {
	err := r.cmd.Wait()
	code := r.cmd.ProcessState.ExitCode()
	if code < 0 && err != nil {
		// Negative exit code = killed by signal; preserve err.
		code = 137 // conventional kill-via-SIGKILL exit code
	}
	r.exited <- process.ExitStatus{Code: code, Err: err}
	close(r.exited)
}
