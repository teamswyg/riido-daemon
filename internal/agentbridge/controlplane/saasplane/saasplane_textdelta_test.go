package saasplane

import (
	"context"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// Raw text deltas are never forwarded one-line-per-token by the event mapper:
// providers (esp. codex) emit them as tiny token/JSON fragments and surfacing
// each as its own progress line produced fragmented, incoherent output. The
// daemon instead accumulates them and forwards an evolving body (see
// TestReportEventCoalescesTextDeltaIntoEvolvingBody).
func TestEventRequestFromAgentEventDropsTextDelta(t *testing.T) {
	assignment := assignmentcontract.Assignment{ID: "asn-1", TaskID: "task-1"}
	if _, ok := eventRequestFromAgentEvent(assignment, agentbridge.Event{
		Kind: agentbridge.EventTextDelta,
		Text: "hello world",
	}); ok {
		t.Fatal("text delta must not be forwarded as a per-delta progress event")
	}

	// Structured progress is still forwarded as a RiidoLog line.
	req, ok := eventRequestFromAgentEvent(assignment, agentbridge.Event{
		Kind: agentbridge.EventProgress,
		Text: "작업을 시작했어요.",
	})
	if !ok || req.EventType != assignmentcontract.EventRiidoLog {
		t.Fatalf("structured progress should still be forwarded: ok=%v req=%+v", ok, req)
	}
}

// ReportEvent coalesces text deltas into an evolving "body so far" and forwards
// it as one tagged progress line (sentinel code 9001 + key assistant.partial)
// carrying the FULL accumulated text — never the raw fragment alone.
func TestReportEventCoalescesTextDeltaIntoEvolvingBody(t *testing.T) {
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

	// First delta flushes immediately (no prior flush); the second crosses the
	// character threshold and flushes the full accumulated body.
	for _, delta := range []string{"Hello, ", "World! This is a streamed answer body."} {
		if err := plane.ReportEvent(context.Background(), req.ID, agentbridge.Event{
			Kind: agentbridge.EventTextDelta,
			Text: delta,
		}); err != nil {
			t.Fatalf("ReportEvent text delta: %v", err)
		}
	}

	if len(fake.events) == 0 {
		t.Fatal("expected at least one forwarded partial-body event")
	}
	last := fake.events[len(fake.events)-1]
	if last.EventType != assignmentcontract.EventRiidoLog {
		t.Fatalf("partial body should be a RiidoLog event, got %q", last.EventType)
	}
	if last.Message != "Hello, World! This is a streamed answer body." {
		t.Fatalf("partial body should carry the full accumulated text, got %q", last.Message)
	}
	if last.Metadata[agentbridge.ProgressMessageMetadataCode] != "9001" ||
		last.Metadata[agentbridge.ProgressMessageMetadataKey] != "assistant.partial" {
		t.Fatalf("partial body must be tagged with sentinel code + key, metadata = %+v", last.Metadata)
	}
}

// On completion with no provider Output (e.g. codex completing via
// thread/status/changed), the accumulated streamed body becomes the terminal
// message so the completed thread shows the actual answer, not an empty/status
// message.
func TestCompleteTaskFallsBackToAccumulatedBody(t *testing.T) {
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
	for _, delta := range []string{"The answer ", "is fully streamed here."} {
		if err := plane.ReportEvent(context.Background(), req.ID, agentbridge.Event{
			Kind: agentbridge.EventTextDelta,
			Text: delta,
		}); err != nil {
			t.Fatalf("ReportEvent: %v", err)
		}
	}
	// Completed with empty Output.
	if err := plane.CompleteTask(context.Background(), req.ID, agentbridge.Result{
		Status: agentbridge.ResultCompleted,
	}); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}

	last := fake.events[len(fake.events)-1]
	if last.State != assignmentcontract.AssignmentCompleted {
		t.Fatalf("expected completed state, got %q", last.State)
	}
	if last.Message != "The answer is fully streamed here." {
		t.Fatalf("completion message should fall back to accumulated body, got %q", last.Message)
	}
}
