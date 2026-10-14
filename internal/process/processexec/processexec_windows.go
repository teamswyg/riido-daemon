//go:build windows

package processexec

import "os/exec"

func configureCommand(*exec.Cmd) {}

func gracefulTerminateCommand(cmd *exec.Cmd) {
	forceTerminateCommand(cmd)
}

func forceTerminateCommand(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}

func normalizeExitCode(code int, err error) int {
	if code >= 0 || err == nil {
		return code
	}
	return 1
}
