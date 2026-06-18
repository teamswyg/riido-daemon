package main

import "testing"

type daemonHealthPayload struct {
	Health string `json:"health"`
}

func TestDaemonHealthEndpoint(t *testing.T) {
	daemon := startForegroundDaemonForStatus(t)
	defer assertForegroundDaemonExits(t, daemon.cancel, daemon.errCh)

	out := daemonEndpointOutput(t, daemon.socket, daemonCommandHealth)
	health := decodeDaemonEndpointJSON[daemonHealthPayload](t, out)
	if health.Health != "ok" {
		t.Fatalf("health: %q", health.Health)
	}
}
