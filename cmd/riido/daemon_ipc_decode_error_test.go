package main

import (
	"encoding/json"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/logging"
)

func TestDaemonIPCReportsMalformedRequest(t *testing.T) {
	server, client := net.Pipe()
	done := make(chan struct{})
	go func() {
		defer close(done)
		log := logging.NewWriterLogger(io.Discard)
		handleDaemonConn(server, startFlags{}, daemonSettings{}, time.Now(), nil, nil, log)
	}()
	t.Cleanup(func() { _ = client.Close() })
	_ = client.SetDeadline(time.Now().Add(time.Second))

	if _, err := client.Write([]byte("not-json\n")); err != nil {
		t.Fatalf("write malformed request: %v", err)
	}
	var response map[string]string
	if err := json.NewDecoder(client).Decode(&response); err != nil {
		t.Fatalf("decode malformed request response: %v", err)
	}
	if response["error"] != "decode request" {
		t.Fatalf("response = %+v, want decode request error", response)
	}
	if !strings.Contains(response["detail"], "invalid character") {
		t.Fatalf("detail = %q, want decoder evidence", response["detail"])
	}
	<-done
}
