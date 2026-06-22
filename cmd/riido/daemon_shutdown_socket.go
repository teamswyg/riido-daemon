package main

import (
	"encoding/json"
	"io"
	"net"
	"time"

	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func tryShutdownViaSocket(socket string, timeout time.Duration, level lifecycle.ShutdownLevel) bool {
	conn, err := net.DialTimeout("unix", socket, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.SetDeadline(time.Now().Add(1 * time.Second))
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
	return waitDaemonSocketClosed(socket, timeout)
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
