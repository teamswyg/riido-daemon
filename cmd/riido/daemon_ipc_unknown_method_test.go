package main

import (
	"encoding/json"
	"io"
	"net"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/logging"
)

func TestDaemonIPCUnknownMethodIncludesSchemaVersion(t *testing.T) {
	server, client := net.Pipe()
	done := make(chan struct{})
	go func() {
		defer close(done)
		log := logging.NewWriterLogger(io.Discard)
		handleDaemonConn(server, startFlags{}, daemonSettings{}, time.Now(), nil, nil, log)
	}()
	t.Cleanup(func() { _ = client.Close() })
	_ = client.SetDeadline(time.Now().Add(time.Second))

	req := daemonRequest{Method: daemonMethod("unknown")}
	if err := json.NewEncoder(client).Encode(req); err != nil {
		t.Fatalf("encode unknown method request: %v", err)
	}
	var response map[string]string
	if err := json.NewDecoder(client).Decode(&response); err != nil {
		t.Fatalf("decode unknown method response: %v", err)
	}
	if response["schema_version"] != DaemonStatusSchemaVersion {
		t.Fatalf("response = %+v, want schema version", response)
	}
	if response["error"] != "unknown method" || response["method"] != "unknown" {
		t.Fatalf("response = %+v, want unknown method evidence", response)
	}
	<-done
}
