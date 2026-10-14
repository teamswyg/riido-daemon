//go:build !windows

package main

import (
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
