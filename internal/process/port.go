// Package process is the Go process port for the run-scope adapter.
//
// The port shields the agentbridge package and the (future) session
// actor from os/exec specifics. A Process produces a RunningProcess
// whose stdout, stderr, and exit signals come over channels rather
// than synchronous reads, so the session actor can select over them
// alongside its mailbox.
//
// The package contains the port + a deterministic Fake. The os/exec
// implementation lives in internal/process/processexec.
package process

import "context"

const (
	// DefaultStdoutBuffer is the bounded process stdout chunk channel size
	// fixed by docs/20-domain/provider-runtime.md §7.5.
	DefaultStdoutBuffer = 64
	// DefaultStderrBuffer is the bounded process stderr chunk channel size
	// fixed by docs/20-domain/provider-runtime.md §7.5.
	DefaultStderrBuffer = 64
)

// Command is the spawn input.
type Command struct {
	Executable string
	Args       []string
	// Env entries are process-specific overrides layered on top of the
	// parent environment by concrete process adapters.
	Env []string
	Dir string
}

// ExitStatus carries the terminal signal from a RunningProcess.
type ExitStatus struct {
	Code int
	Err  error
}

// Process spawns a RunningProcess. Implementations MUST return a
// process with closed stdout/stderr channels once Exited has been
// signaled, so consumers can use `range` to drain.
type Process interface {
	Start(ctx context.Context, cmd Command) (RunningProcess, error)
}

// RunningProcess exposes the streams and control surface of a single
// spawned process. All methods are safe to call from one goroutine;
// the session actor is the only intended caller.
type RunningProcess interface {
	Stdout() <-chan []byte
	Stderr() <-chan []byte
	Exited() <-chan ExitStatus
	WriteStdin([]byte) error
	CloseStdin() error
	Kill(context.Context) error
}
