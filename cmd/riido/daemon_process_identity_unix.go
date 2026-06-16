//go:build !windows

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

func validateDaemonProcessIdentity(pid int) error {
	if pid <= 0 {
		return daemonErrorf(ErrDaemonProcess, "stop.verify-pid", "invalid daemon pid %d", pid)
	}
	command, err := daemonProcessCommandLine(pid)
	if err != nil {
		return daemonWrapf(ErrDaemonProcess, "stop.verify-pid", err, "verify daemon process %d", pid)
	}
	if daemonCommandLineLooksLikeSelf(command) {
		return nil
	}
	return daemonErrorf(ErrDaemonProcess, "stop.verify-pid", "pid %d is not a riido daemon foreground process", pid)
}

func daemonProcessCommandLine(pid int) (string, error) {
	var lastErr error
	for _, ps := range []string{"/bin/ps", "/usr/bin/ps", "ps"} {
		out, err := exec.Command(ps, "-p", strconv.Itoa(pid), "-o", "command=").Output()
		if err == nil {
			return strings.TrimSpace(string(bytes.TrimSpace(out))), nil
		}
		lastErr = err
	}
	return "", lastErr
}

func daemonCommandLineLooksLikeSelf(command string) bool {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return false
	}
	if !daemonCommandBinaryLooksLikeSelf(fields[0]) {
		return false
	}
	for i, field := range fields {
		if field != string(mainCommandDaemon) {
			continue
		}
		if i+1 < len(fields) && fields[i+1] == string(daemonCommandStart) && fieldsAfterContain(fields[i+2:], "--foreground") {
			return true
		}
	}
	return false
}

func daemonCommandBinaryLooksLikeSelf(argv0 string) bool {
	current, err := os.Executable()
	if err != nil {
		return false
	}
	binaryName := filepath.Base(argv0)
	currentName := filepath.Base(current)
	return binaryName == currentName || binaryName == "riido"
}

func fieldsAfterContain(fields []string, want string) bool {
	return slices.Contains(fields, want)
}
