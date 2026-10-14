//go:build windows

package main

func validateDaemonProcessIdentity(pid int, _ daemonPIDIdentity) error {
	return daemonErrorf(ErrDaemonProcess, "stop.verify-pid", "pid fallback is not supported on windows without daemon process identity evidence: %d", pid)
}
