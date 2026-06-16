//go:build !windows

package main

import "testing"

func TestDaemonCommandLineMatchesPIDIdentityRequiresSocket(t *testing.T) {
	command := "/tmp/riido daemon start --foreground --socket /tmp/riido-a.sock --pid-file /tmp/riido.pid"
	identity := daemonPIDIdentity{
		SchemaVersion: daemonPIDIdentitySchemaVersion,
		PID:           1234,
		Socket:        "/tmp/riido-a.sock",
	}
	if !daemonCommandLineMatchesPIDIdentity(command, identity) {
		t.Fatalf("expected command line to match pid identity")
	}

	identity.Socket = "/tmp/riido-b.sock"
	if daemonCommandLineMatchesPIDIdentity(command, identity) {
		t.Fatalf("pid identity for another socket must not match")
	}

	identity.Socket = "/tmp/riido-a.sock"
	if daemonCommandLineMatchesPIDIdentity("/tmp/riido daemon start --socket /tmp/riido-a.sock", identity) {
		t.Fatalf("pid identity must require daemon start --foreground")
	}
}
