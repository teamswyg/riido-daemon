//go:build !windows

package main

import "testing"

func TestDaemonCommandLineMatchesPIDIdentityRequiresSocket(t *testing.T) {
	command := "/tmp/riido daemon start --foreground --socket /tmp/riido-a.sock --pid-file /tmp/riido.pid"
	identity := daemonPIDIdentity{
		SchemaVersion: daemonPIDIdentitySchemaVersion,
		PID:           1234,
		Executable:    "/tmp/riido",
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
	if daemonCommandLineMatchesPIDIdentity("/tmp/riido daemon start --foreground --socket /tmp/riido-a.sock.bak", identity) {
		t.Fatalf("pid identity socket must match an exact argv field")
	}
	if !daemonCommandLineMatchesPIDIdentity("/tmp/riido daemon start --foreground --socket=/tmp/riido-a.sock", identity) {
		t.Fatalf("expected --socket=value form to match pid identity")
	}
}

func TestDaemonCommandLineMatchesPIDIdentityRequiresExecutable(t *testing.T) {
	identity := daemonPIDIdentity{
		SchemaVersion: daemonPIDIdentitySchemaVersion,
		PID:           1234,
		Executable:    "/opt/riido/bin/riido",
		Socket:        "/tmp/riido.sock",
	}

	if daemonCommandLineMatchesPIDIdentity("/tmp/riido daemon start --foreground --socket /tmp/riido.sock", identity) {
		t.Fatal("pid identity executable mismatch must not match")
	}
	if !daemonCommandLineMatchesPIDIdentity("/opt/riido/bin/riido daemon start --foreground --socket /tmp/riido.sock", identity) {
		t.Fatal("pid identity executable should match exact argv0 path")
	}
}
