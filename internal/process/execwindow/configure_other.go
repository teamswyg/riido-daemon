//go:build !windows

package execwindow

import "os/exec"

// Hide is a no-op outside Windows, where child processes do not create a
// separate console window by default.
func Hide(*exec.Cmd) {}
