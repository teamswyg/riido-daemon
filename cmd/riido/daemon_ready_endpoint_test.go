package main

import "testing"

type daemonReadyPayload struct {
	Ready             bool   `json:"ready"`
	Readiness         string `json:"readiness"`
	RuntimeCount      int    `json:"runtime_count"`
	RuntimeResponding int    `json:"runtime_responding"`
	SchemaVersion     string `json:"schema_version"`
}

func TestDaemonReadyEndpoint(t *testing.T) {
	daemon := startForegroundDaemonForStatus(t)
	defer assertForegroundDaemonExits(t, daemon.cancel, daemon.errCh)

	out := daemonEndpointOutput(t, daemon.socket, daemonCommandReady)
	ready := decodeDaemonEndpointJSON[daemonReadyPayload](t, out)
	if !ready.Ready || ready.Readiness != "ready" || ready.RuntimeCount != 4 || ready.RuntimeResponding != 4 {
		t.Fatalf("ready payload mismatch: %+v\n%s", ready, out)
	}
}
