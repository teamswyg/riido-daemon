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

// daemonPIDProbablyAlive is conservative on Windows: process liveness cannot be
// probed reliably here yet, so it always reports alive. This prevents a stale
// ".claim" reclaim from spawning a second daemon while the first is still live
// (a reliable Windows liveness probe is tracked separately as D7).
func daemonPIDProbablyAlive(int) bool {
	return true
}
