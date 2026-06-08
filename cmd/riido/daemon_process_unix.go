//go:build !windows

package main

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

func setDaemonChildSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func daemonInterruptSignals() []os.Signal {
	return []os.Signal{syscall.SIGTERM, syscall.SIGINT}
}

func signalDaemonProcessTerm(proc *os.Process) error {
	return proc.Signal(syscall.SIGTERM)
}

func signalDaemonProcessKill(proc *os.Process) error {
	return proc.Signal(syscall.SIGKILL)
}

func daemonProcessExists(proc *os.Process) bool {
	return proc.Signal(syscall.Signal(0)) == nil
}

// daemonPIDProbablyAlive reports whether pid is a live process. It is
// conservative: only a definitive ESRCH ("no such process") is treated as dead;
// any ambiguous result (EPERM, lookup error) is treated as alive so a stale-lock
// reclaim never steals a lock from a process we merely cannot signal.
func daemonPIDProbablyAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return true
	}
	switch err := proc.Signal(syscall.Signal(0)); {
	case err == nil:
		return true
	case errors.Is(err, syscall.ESRCH), errors.Is(err, os.ErrProcessDone):
		// ESRCH: no such PID (a foreign dead owner). ErrProcessDone: a child this
		// process already reaped. Both mean the recorded owner is gone.
		return false
	default:
		// EPERM (alive, not ours) or any ambiguous error: assume alive so a
		// stale-lock reclaim never steals a lock we cannot prove is free.
		return true
	}
}
