package saasplane

import (
	"context"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestPlaneReportsStructuredProgressMetadata(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-1",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "ship it",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-1",
	})
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	if err := plane.ReportEvent(context.Background(), req.ID, agentbridge.Event{
		Kind:         agentbridge.EventProgress,
		Text:         "생각 중. . .",
		ProgressCode: agentbridge.ProgressCodeAgentThinking,
		ProgressKey:  "agent.thinking",
	}); err != nil {
		t.Fatalf("ReportEvent: %v", err)
	}
	if len(fake.events) != 1 {
		t.Fatalf("events = %+v", fake.events)
	}
	event := fake.events[0]
	if event.Metadata[agentbridge.ProgressMessageMetadataCode] != "1001" ||
		event.Metadata[agentbridge.ProgressMessageMetadataKey] != "agent.thinking" {
		t.Fatalf("metadata = %+v", event.Metadata)
	}
}
