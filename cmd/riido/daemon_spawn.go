package main

import (
	"os"
	"os/exec"
)

// daemonSpawnHelper builds the exec.Cmd used by the background wrapper. Tests
// override it before concurrent execution begins.
var daemonSpawnHelper = defaultDaemonSpawnHelper

func defaultDaemonSpawnHelper(args []string) (*exec.Cmd, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, daemonWrapf(ErrDaemonProcess, "spawn.locate-executable", err, "locate daemon binary")
	}
	return exec.Command(exe, args...), nil
}
