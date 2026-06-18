package saasplane

import (
	"context"
	"strings"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func reportSameTaskHeartbeat(t *testing.T, fake *fakeAssignmentServer, plane *Plane, ids ...string) {
	t.Helper()
	heartbeat := controlplane.RuntimeHeartbeat{
		RuntimeID:      RuntimeIDForAgent("daemon-1", lifecycleAgent()),
		RunningTaskIDs: ids,
	}
	if err := plane.Heartbeat(context.Background(), heartbeat); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	heartbeats := fake.heartbeatsFor("jykim1")
	if len(heartbeats) != 1 || strings.Join(heartbeats[0].ActiveAssignmentIDs, ",") != "asn-1,asn-2" {
		t.Fatalf("heartbeats = %+v", heartbeats)
	}
}

func reportSameTaskProgress(
	t *testing.T,
	fake *fakeAssignmentServer,
	plane *Plane,
	firstID string,
	secondID string,
	second assignmentcontract.Assignment,
) {
	t.Helper()
	reportLifecycleProgress(t, plane, firstID, "first progress")
	reportLifecycleProgress(t, plane, secondID, "second progress")
	last := fake.events[len(fake.events)-1]
	if last.AssignmentID != second.ID || last.TaskID != second.TaskID {
		t.Fatalf("second event identity = %+v", last)
	}
}

func reportLifecycleProgress(t *testing.T, plane *Plane, assignmentID, text string) {
	t.Helper()
	event := agentbridge.Event{Kind: agentbridge.EventProgress, Text: text}
	if err := plane.ReportEvent(context.Background(), assignmentID, event); err != nil {
		t.Fatalf("ReportEvent %s: %v", assignmentID, err)
	}
}

func lifecycleAgent() AgentBinding {
	return AgentBinding{AgentID: "jykim1", RuntimeProvider: "codex"}
}
