package main

import (
	"github.com/teamswyg/riido-daemon/pkg/failure"
)

const daemonErrorLayer failure.Layer = "daemon"

var (
	ErrDaemonUsage        = failure.NewSentinel(daemonErrorLayer, "usage")
	ErrDaemonConfig       = failure.NewSentinel(daemonErrorLayer, "config")
	ErrDaemonIO           = failure.NewSentinel(daemonErrorLayer, "io")
	ErrDaemonLock         = failure.NewSentinel(daemonErrorLayer, "lock")
	ErrDaemonSocket       = failure.NewSentinel(daemonErrorLayer, "socket")
	ErrDaemonProcess      = failure.NewSentinel(daemonErrorLayer, "process")
	ErrDaemonRuntime      = failure.NewSentinel(daemonErrorLayer, "runtime")
	ErrDaemonSupervisor   = failure.NewSentinel(daemonErrorLayer, "supervisor")
	ErrDaemonControlPlane = failure.NewSentinel(daemonErrorLayer, "control-plane")
)

func daemonErrorf(kind failure.Sentinel, op, format string, args ...any) error {
	return failure.New(kind, op, failure.Format(format, args...))
}

func daemonWrapf(kind failure.Sentinel, op string, cause error, format string, args ...any) error {
	return failure.Wrap(kind, op, failure.Format(format, args...), cause)
}
