//go:build !windows

package main

import (
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
