//go:build windows

package main

import (
	"os"
	"os/exec"

	"github.com/teamswyg/riido-daemon/internal/process/execwindow"
)

func setDaemonChildSysProcAttr(cmd *exec.Cmd) {
	execwindow.Hide(cmd)
}

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
