//go:build windows

package main

import (
	"os"
	"os/exec"
)

func setDaemonChildSysProcAttr(*exec.Cmd) {}

func daemonInterruptSignals() []os.Signal {
	return []os.Signal{os.Interrupt}
}

func signalDaemonProcessTerm(proc *os.Process) error {
	return proc.Kill()
}

func signalDaemonProcessKill(proc *os.Process) error {
	return proc.Kill()
}

func daemonProcessExists(*os.Process) bool {
	return false
}
