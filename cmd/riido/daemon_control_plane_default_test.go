package main

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestBuildDaemonControlPlaneUsesMemoryByDefault(t *testing.T) {
	source, reporter, kind, err := buildDaemonControlPlane(daemonSettings{}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if kind != "memory" {
		t.Fatalf("kind = %q", kind)
	}
	if _, ok := source.(*controlplane.MemorySource); !ok {
		t.Fatalf("source type = %T", source)
	}
	if _, ok := reporter.(*controlplane.MemoryReporter); !ok {
		t.Fatalf("reporter type = %T", reporter)
	}
}
