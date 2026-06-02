//go:build windows

package processexec

import "os/exec"

func configureCommand(*exec.Cmd) {}

func terminateCommand(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}
