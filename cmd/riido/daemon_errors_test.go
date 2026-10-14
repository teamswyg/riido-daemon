package main

import (
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/pkg/failure"
)

func TestDaemonUsageErrorIsClassified(t *testing.T) {
	_, err := parseStartFlags([]string{"--socket"})
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !errors.Is(err, ErrDaemonUsage) {
		t.Fatalf("errors.Is(err, ErrDaemonUsage) = false for %v", err)
	}

	operational, ok := failure.AsOperational(err)
	if !ok {
		t.Fatal("expected operational error")
	}
	if operational.Op() != "start.parse-flags" {
		t.Fatalf("Op() = %q, want %q", operational.Op(), "start.parse-flags")
	}
}

func TestDaemonControlPlaneConfigErrorIsClassified(t *testing.T) {
	_, _, _, err := buildDaemonControlPlane(daemonSettings{
		TaskDBSourcePath: "task-db.json",
		TaskQueueDir:     "queue",
	}, time.Now())
	if err == nil {
		t.Fatal("expected config error")
	}
	if !errors.Is(err, ErrDaemonConfig) {
		t.Fatalf("errors.Is(err, ErrDaemonConfig) = false for %v", err)
	}
}
