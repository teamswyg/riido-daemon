package main

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/mwsdbridge"
)

func serveTestMwsd(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("/tmp", "riido-mwsd-test-")
	if err != nil {
		t.Fatalf("create short socket dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	socketPath := filepath.Join(dir, "mwsd.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen unix: %v", err)
	}
	done := make(chan struct{})
	snapshot := cliMwsdSnapshot()
	go serveTestMwsdLoop(listener, snapshot, done)
	return socketPath, func() {
		_ = listener.Close()
		waitForMwsdServerStop(t, done)
	}
}

func serveTestMwsdLoop(listener net.Listener, snapshot mwsdbridge.Snapshot, done chan<- struct{}) {
	defer close(done)
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go serveTestMwsdConn(conn, snapshot)
	}
}

func waitForMwsdServerStop(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for mwsd test server shutdown")
	}
}

func serveTestMwsdConn(conn net.Conn, snapshot mwsdbridge.Snapshot) {
	defer conn.Close()
	method, ok := readMwsdTestMethod(conn)
	if !ok {
		return
	}
	data, ok := mwsdTestResponse(snapshot, method)
	if !ok {
		_ = json.NewEncoder(conn).Encode(map[string]any{"ok": false, "method": method, "error": "unknown method"})
		return
	}
	writeMwsdTestResponse(conn, method, data)
}
