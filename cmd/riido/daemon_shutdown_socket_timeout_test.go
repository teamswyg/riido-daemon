package main

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func TestTryShutdownViaSocketHonorsTimeoutBudgetWhenAckStalls(t *testing.T) {
	sock := daemonSocketPath(t)
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	release := make(chan struct{})
	defer close(release)

	go serveStalledShutdownAck(ln, release)

	start := time.Now()
	if ok := tryShutdownViaSocket(sock, 50*time.Millisecond, lifecycle.ShutdownForced); ok {
		t.Fatal("stalled shutdown ack should not complete")
	}
	if elapsed := time.Since(start); elapsed > 300*time.Millisecond {
		t.Fatalf("shutdown socket timeout took %s, want under 300ms", elapsed)
	}
}

func serveStalledShutdownAck(ln net.Listener, release <-chan struct{}) {
	conn, err := ln.Accept()
	if err != nil {
		return
	}
	defer conn.Close()
	var req daemonRequest
	_ = json.NewDecoder(conn).Decode(&req)
	<-release
}
