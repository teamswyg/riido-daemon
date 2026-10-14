//go:build windows

package execwindow

import (
	"os/exec"
	"testing"
)

func TestHidePreventsWindowsConsoleCreation(t *testing.T) {
	cmd := exec.Command("cmd.exe", "/c", "exit", "0")
	Hide(cmd)
	if cmd.SysProcAttr == nil || !cmd.SysProcAttr.HideWindow {
		t.Fatal("Windows child command must hide its window")
	}
	if cmd.SysProcAttr.CreationFlags&createNoWindow == 0 {
		t.Fatal("Windows child command must use CREATE_NO_WINDOW")
	}
}
