package main

import (
	"encoding/json"
	"io"
	"net"
	"time"

	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func tryShutdownViaSocket(socket string, timeout time.Duration, level lifecycle.ShutdownLevel) bool {
	if timeout <= 0 {
		return false
	}
	deadline := time.Now().Add(timeout)
	conn, err := net.DialTimeout("unix", socket, shutdownSocketDialTimeout(timeout))
	if err != nil {
		return false
	}
	if err := conn.SetDeadline(deadline); err != nil {
		_ = conn.Close()
		return false
	}
	if err := json.NewEncoder(conn).Encode(daemonRequest{
		Method:        daemonMethodShutdown,
		ShutdownLevel: level.String(),
		Force:         level.IsForced(),
	}); err != nil {
		_ = conn.Close()
		return false
	}
	_, _ = io.ReadAll(conn)
	_ = conn.Close()
	return waitDaemonSocketClosedUntil(socket, deadline)
}

func waitDaemonSocketClosed(socket string, timeout time.Duration) bool {
	return waitDaemonShutdownCondition(timeout, func() bool {
		c, err := net.DialTimeout("unix", socket, 100*time.Millisecond)
		if err != nil {
			return true
		}
		_ = c.Close()
		return false
	})
}

func shutdownSocketDialTimeout(timeout time.Duration) time.Duration {
	return min(timeout, 500*time.Millisecond)
}

func waitDaemonSocketClosedUntil(socket string, deadline time.Time) bool {
	remaining := time.Until(deadline)
	if remaining <= 0 {
		return false
	}
	return waitDaemonSocketClosed(socket, remaining)
}
