package main

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func TestTryShutdownViaSocketSendsForcedLevel(t *testing.T) {
	sock := daemonSocketPath(t)
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	received := make(chan daemonRequest, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		var req daemonRequest
		_ = json.NewDecoder(conn).Decode(&req)
		received <- req
		writeShutdownAck(conn, req.lifecycleShutdownLevel())
		_ = ln.Close()
	}()

	if ok := tryShutdownViaSocket(sock, time.Second, lifecycle.ShutdownForced); !ok {
		t.Fatal("forced shutdown socket request did not complete")
	}
	select {
	case req := <-received:
		if req.Method != daemonMethodShutdown || !req.Force || req.ShutdownLevel != lifecycle.ShutdownForced.String() {
			t.Fatalf("shutdown request = %+v", req)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown request was not received")
	}
}
