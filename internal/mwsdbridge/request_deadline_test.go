package mwsdbridge

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRequestReadHonorsClientTimeout(t *testing.T) {
	dir, err := os.MkdirTemp("/tmp", "riido-mwsd-timeout-")
	if err != nil {
		t.Fatalf("create socket dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	socketPath := filepath.Join(dir, "mwsd.sock")
	stop := serveStalledMwsd(t, socketPath)
	defer stop()

	client := NewClient(socketPath)
	client.Timeout = 50 * time.Millisecond

	var status Status
	started := time.Now()
	err = client.Request(context.Background(), "status", &status)
	if err == nil {
		t.Fatal("Request returned nil error for stalled mwsd response")
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("Request ignored timeout: elapsed=%s err=%v", elapsed, err)
	}
	if !strings.Contains(err.Error(), "read mwsd response") {
		t.Fatalf("error = %q, want read evidence", err.Error())
	}
}

func serveStalledMwsd(t *testing.T, socketPath string) func() {
	t.Helper()
	_ = os.Remove(socketPath)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen unix socket: %v", err)
	}
	release := make(chan struct{})
	done := make(chan struct{})
	go serveStalledMwsdLoop(listener, release, done)
	return func() {
		close(release)
		_ = listener.Close()
		<-done
		_ = os.Remove(socketPath)
	}
}

func serveStalledMwsdLoop(listener net.Listener, release <-chan struct{}, done chan<- struct{}) {
	defer close(done)
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go stallMwsdConn(conn, release)
	}
}

func stallMwsdConn(conn net.Conn, release <-chan struct{}) {
	defer conn.Close()
	var req request
	_ = json.NewDecoder(conn).Decode(&req)
	<-release
}
