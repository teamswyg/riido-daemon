package main

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/saasplane"
)

func TestBuildDaemonControlPlaneUsesSaaS(t *testing.T) {
	source, reporter, kind, err := buildDaemonControlPlane(daemonSettings{
		DaemonID:     "daemon-1",
		DeviceName:   "device-1",
		SaaSURL:      "http://127.0.0.1:1",
		DeviceID:     "device-1",
		DeviceSecret: "rdev-secret",
	}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if kind != "saas" {
		t.Fatalf("kind = %q", kind)
	}
	plane, ok := source.(*saasplane.Plane)
	if !ok {
		t.Fatalf("source type = %T", source)
	}
	defer plane.Close()
	if _, ok := reporter.(*saasplane.Plane); !ok {
		t.Fatalf("reporter type = %T", reporter)
	}
}
