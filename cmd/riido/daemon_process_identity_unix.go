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

func validateDaemonProcessIdentity(pid int, identity daemonPIDIdentity) error {
	if pid <= 0 {
		return daemonErrorf(ErrDaemonProcess, "stop.verify-pid", "invalid daemon pid %d", pid)
	}
	if identity.SchemaVersion != daemonPIDIdentitySchemaVersion {
		return daemonErrorf(ErrDaemonProcess, "stop.verify-pid", "unsupported daemon pid identity schema_version %q", identity.SchemaVersion)
	}
	if identity.PID != pid {
		return daemonErrorf(ErrDaemonProcess, "stop.verify-pid", "daemon pid identity pid %d does not match pid file %d", identity.PID, pid)
	}
	if strings.TrimSpace(identity.Socket) == "" {
		return daemonErrorf(ErrDaemonProcess, "stop.verify-pid", "daemon pid identity socket is required")
	}
	command, err := daemonProcessCommandLine(pid)
	if err != nil {
		return daemonWrapf(ErrDaemonProcess, "stop.verify-pid", err, "verify daemon process %d", pid)
	}
	if daemonCommandLineMatchesPIDIdentity(command, identity) {
		return nil
	}
	return daemonErrorf(ErrDaemonProcess, "stop.verify-pid", "pid %d is not the riido daemon foreground process described by pid identity", pid)
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

func daemonCommandLineMatchesPIDIdentity(command string, identity daemonPIDIdentity) bool {
	fields := strings.Fields(command)
	if len(fields) == 0 || !daemonCommandBinaryMatchesPIDIdentity(fields[0], identity) {
		return false
	}
	if !daemonCommandLineLooksLikeForegroundStart(fields) {
		return false
	}
	socket := strings.TrimSpace(identity.Socket)
	if socket == "" {
		return false
	}
	return daemonCommandFieldsContainSocket(fields, socket)
}

func daemonCommandBinaryMatchesPIDIdentity(argv0 string, identity daemonPIDIdentity) bool {
	expected := strings.TrimSpace(identity.Executable)
	if expected == "" {
		return daemonCommandBinaryLooksLikeSelf(argv0)
	}
	return filepath.Clean(argv0) == filepath.Clean(expected)
}

func daemonCommandLineLooksLikeForegroundStart(fields []string) bool {
	if len(fields) == 0 {
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

func daemonCommandFieldsContainSocket(fields []string, socket string) bool {
	for i, field := range fields {
		if field == "--socket" {
			return i+1 < len(fields) && fields[i+1] == socket
		}
		if value, ok := strings.CutPrefix(field, "--socket="); ok {
			return value == socket
		}
	}
	return false
}
