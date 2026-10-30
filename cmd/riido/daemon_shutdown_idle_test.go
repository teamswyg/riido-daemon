package main

import (
	"encoding/json"
	"net"
	"testing"
	"time"
)

func TestIdleShutdownLegacyAndBusyNeverFallsBackToPID(t *testing.T) {
	for _, response := range []map[string]string{{"error": "unknown method"}, {"shutdown": "deferred"}} {
		sock := daemonSocketPath(t)
		ln, err := net.Listen("unix", sock)
		if err != nil {
			t.Fatal(err)
		}
		done := make(chan struct{})
		go func() {
			defer close(done)
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			defer conn.Close()
			var req daemonRequest
			json.NewDecoder(conn).Decode(&req)
			if req.Method != "shutdown-if-idle" {
				t.Errorf("unsafe method %s", req.Method)
			}
			json.NewEncoder(conn).Encode(response)
		}()
		// A bogus PID file would produce a different error if fallback were attempted.
		err = stopDaemonWithFlags(daemonStopFlags{socket: sock, pidFile: "/does-not-exist", timeoutSeconds: 1})
		if err == nil {
			t.Fatal("unsafe shutdown accepted")
		}
		conn, dialErr := net.DialTimeout("unix", sock, time.Second)
		if dialErr != nil {
			t.Fatal("legacy listener stopped")
		}
		conn.Close()
		ln.Close()
		<-done
	}
}
