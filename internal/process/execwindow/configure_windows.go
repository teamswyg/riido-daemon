//go:build windows

package execwindow

import (
	"os/exec"
	"syscall"
)

const createNoWindow = 0x08000000

// Hide keeps daemon-owned child commands in the daemon's background session.
// HideWindow covers console applications while CREATE_NO_WINDOW also prevents
// .cmd/.bat shims from flashing a transient Command Prompt window.
func Hide(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
	cmd.SysProcAttr.CreationFlags |= createNoWindow
}
