package main

import (
	"os"
	"os/exec"
)

func daemonBackgroundChildArgs(flags startFlags) []string {
	childArgs := []string{"daemon", "start", "--foreground", "--socket", flags.socket}
	if flags.pidFile != "" {
		childArgs = append(childArgs, "--pid-file", flags.pidFile)
	}
	if flags.logFile != "" {
		childArgs = append(childArgs, "--log-file", flags.logFile)
	}
	if flags.lockFile != "" {
		childArgs = append(childArgs, "--lock-file", flags.lockFile)
	}
	return childArgs
}

func prepareDaemonBackgroundChild(flags startFlags) (*exec.Cmd, *os.File, error) {
	cmd, err := daemonSpawnHelper(daemonBackgroundChildArgs(flags))
	if err != nil {
		return nil, nil, err
	}
	devNull, err := openDaemonChildDevNull()
	if err != nil {
		return nil, nil, err
	}
	cmd.Stdin = devNull
	cmd.Stdout = devNull
	cmd.Stderr = devNull
	setDaemonChildSysProcAttr(cmd)
	return cmd, devNull, nil
}

func openDaemonChildDevNull() (*os.File, error) {
	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return nil, daemonWrapf(ErrDaemonIO, "background.open-dev-null", err, "open /dev/null")
	}
	return devNull, nil
}
