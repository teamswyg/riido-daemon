package main

import (
	"encoding/json"
	"io"
	"net"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func TestDaemonShutdownRequestCarriesForcedLevel(t *testing.T) {
	server, client := net.Pipe()
	shutdownCh := make(chan lifecycle.ShutdownLevel, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		handleDaemonConn(server, startFlags{}, daemonSettings{}, time.Now(), nil, nil, nil, shutdownCh, logging.NewWriterLogger(io.Discard))
	}()
	t.Cleanup(func() { _ = client.Close() })
	_ = client.SetDeadline(time.Now().Add(time.Second))

	req := daemonRequest{Method: daemonMethodShutdown, ShutdownLevel: "forced"}
	if err := json.NewEncoder(client).Encode(req); err != nil {
		t.Fatalf("encode shutdown request: %v", err)
	}
	var ack map[string]string
	if err := json.NewDecoder(client).Decode(&ack); err != nil {
		t.Fatalf("decode shutdown ack: %v", err)
	}
	if ack["shutdown"] != "accepted" || ack["shutdown_level"] != lifecycle.ShutdownForced.String() {
		t.Fatalf("shutdown ack = %+v", ack)
	}
	select {
	case level := <-shutdownCh:
		if level != lifecycle.ShutdownForced {
			t.Fatalf("shutdown level = %s, want %s", level, lifecycle.ShutdownForced)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown level was not delivered")
	}
	<-done
}
