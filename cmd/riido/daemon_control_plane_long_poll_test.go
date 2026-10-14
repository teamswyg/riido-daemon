package main

import (
	"context"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/saasplane"
)

func TestBuildDaemonControlPlaneSaaSUsesDefaultLongPollWait(t *testing.T) {
	pollSeen := make(chan assignmentcontract.PollRequest, 1)
	server := newDaemonLongPollServer(t, pollSeen)
	source, reporter, kind, err := buildDaemonControlPlane(daemonSettings{
		DaemonID:     "device-1",
		DeviceName:   "device-1",
		SaaSURL:      server.URL,
		DeviceID:     "device-1",
		DeviceSecret: "rdev-secret",
	}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closer, ok := source.(interface{ Close() }); ok {
			closer.Close()
		}
	})
	sourcePlane, sourceOK := source.(*saasplane.Plane)
	reporterPlane, reporterOK := reporter.(*saasplane.Plane)
	if kind != "saas" || !sourceOK || !reporterOK || sourcePlane != reporterPlane {
		t.Fatalf("control plane kind/source/reporter = %q %T %T", kind, source, reporter)
	}
	req, err := source.ClaimTask(context.Background(), "device-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	if req != nil {
		t.Fatalf("empty fake server should not claim task: %+v", req)
	}
	assertDefaultLongPollWait(t, pollSeen)
}
