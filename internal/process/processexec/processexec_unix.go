//go:build !windows

package processexec

import (
	"errors"
	"os/exec"
	"syscall"
)

func configureCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func gracefulTerminateCommand(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	pid := cmd.Process.Pid
	_ = syscall.Kill(-pid, syscall.SIGTERM)
	_ = cmd.Process.Signal(syscall.SIGTERM)
}

func forceTerminateCommand(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	// Best-effort: terminate the process group first so child processes cannot
	// keep stdout/stderr pipes open after the shell exits. Fall back to killing
	// the single PID.
	pid := cmd.Process.Pid
	_ = syscall.Kill(-pid, syscall.SIGKILL)
	_ = cmd.Process.Kill()
}

func normalizeExitCode(code int, err error) int {
	if code >= 0 || err == nil {
		return code
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			return 128 + int(status.Signal())
		}
	}
	return 137
}
