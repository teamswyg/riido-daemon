package main

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

func readDaemonPIDFile(t *testing.T, path string) int {
	t.Helper()
	pidBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read pid file: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil || pid <= 0 {
		t.Fatalf("invalid pid: %q (err=%v)", pidBytes, err)
	}
	return pid
}

func assertBackgroundPIDIdentity(t *testing.T, paths daemonBackgroundPaths, pid int) {
	t.Helper()
	if pid == os.Getpid() {
		t.Fatalf("pid file holds parent's pid (%d); daemon did not detach", pid)
	}
	identity, ok, err := loadDaemonPIDIdentity(paths.pid)
	if err != nil {
		t.Fatalf("read pid identity: %v", err)
	}
	if !ok {
		t.Fatal("pid identity sidecar missing")
	}
	if identity.SchemaVersion != daemonPIDIdentitySchemaVersion {
		t.Fatalf("pid identity schema version = %+v", identity)
	}
	if identity.PID != pid || identity.Socket != paths.socket {
		t.Fatalf("pid identity = %+v, pid=%d socket=%s", identity, pid, paths.socket)
	}
}
