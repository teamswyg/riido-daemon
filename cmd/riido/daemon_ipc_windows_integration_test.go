//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/logging"
)

func TestWindowsNamedPipeDaemonStatusRoundTrip(t *testing.T) {
	pipe := fmt.Sprintf(`\\.\pipe\riido-daemon-status-test-%d`, time.Now().UnixNano())
	listener, cleanup, err := listenDaemonSocket(pipe)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		handleDaemonConn(
			conn,
			startFlags{socket: pipe},
			daemonSettings{DaemonVersion: "windows-ipc-test"},
			time.Now().Add(-3*time.Second),
			nil,
			nil,
			nil,
			nil,
			logging.NewWriterLogger(io.Discard),
		)
	}()

	conn, err := dialDaemonSocket(pipe, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	if err := json.NewEncoder(conn).Encode(daemonRequest{Method: daemonMethodStatus}); err != nil {
		t.Fatal(err)
	}

	body, err := io.ReadAll(conn)
	if err != nil {
		t.Fatalf("read complete status response: %v", err)
	}
	var status daemonStatus
	if err := json.Unmarshal(body, &status); err != nil {
		t.Fatal(err)
	}
	if status.DaemonVersion != "windows-ipc-test" {
		t.Fatalf("daemon version = %q", status.DaemonVersion)
	}
	if status.SocketPath != pipe {
		t.Fatalf("socket path = %q, want %q", status.SocketPath, pipe)
	}
	if status.PID == 0 || status.UptimeSeconds < 2 {
		t.Fatalf("invalid status pid=%d uptime=%d", status.PID, status.UptimeSeconds)
	}

	select {
	case <-serverDone:
	case <-time.After(2 * time.Second):
		t.Fatal("server connection did not close")
	}
}
