package processexec

import (
	"context"
	"errors"
	"os/exec"
	"sync"

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
	command := exec.CommandContext(cmdCtx, cmd.Executable, cmd.Args...)
	command.Env = mergeEnv(cmd.Env)
	command.Dir = cmd.Dir
	configureCommand(command)
	return startRunningProcess(cmdCtx, cancel, command)
}

func startRunningProcess(
	ctx context.Context,
	cancel context.CancelFunc,
	command *exec.Cmd,
) (process.RunningProcess, error) {
	stdinPipe, err := command.StdinPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	running := newExecRunning(command, cancel, stdinPipe, &sync.Mutex{})
	command.Stdout = streamWriter{out: running.stdout}
	command.Stderr = streamWriter{out: running.stderr}
	if err := command.Start(); err != nil {
		cancel()
		return nil, err
	}
	go running.killOnContext(ctx)
	go running.waitExit()
	return running, nil
}
