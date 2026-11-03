package main

import (
	"encoding/json"
	"fmt"
	"time"
)

func shutdownIdleDaemon(socket string, timeout time.Duration) error {
	conn, err := dialDaemonSocket(socket, shutdownSocketDialTimeout(timeout))
	if err != nil {
		return err
	}
	defer conn.Close()
	deadline := time.Now().Add(timeout)
	if err := conn.SetDeadline(deadline); err != nil {
		return err
	}
	if err := json.NewEncoder(conn).Encode(daemonRequest{Method: "shutdown-if-idle"}); err != nil {
		return err
	}
	var response struct {
		Shutdown string `json:"shutdown"`
	}
	if err := json.NewDecoder(conn).Decode(&response); err != nil {
		return err
	}
	if response.Shutdown != "accepted" {
		// Old daemons return unknown-method. Fail closed, including for PID fallback:
		// a desktop updater must never kill a legacy daemon to activate this release.
		return fmt.Errorf("daemon stop deferred: idle shutdown unavailable or execution/delivery pending")
	}
	_ = conn.Close()
	if !waitDaemonSocketClosedUntil(socket, deadline) {
		return fmt.Errorf("idle daemon shutdown is still pending")
	}
	return nil
}
