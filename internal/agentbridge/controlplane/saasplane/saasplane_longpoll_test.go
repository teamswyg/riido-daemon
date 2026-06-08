package saasplane

import (
	"context"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

// The claim poll forwards the long-poll hint (wait_ms) and is bounded by
// LongPollTimeout, NOT the short RequestTimeout that governs heartbeat/events.
// The fake server delays its poll response by 250ms; with RequestTimeout=50ms
// the claim still succeeds because the poll uses the 2s LongPollTimeout.
func TestPlaneClaimSendsWaitMsAndUsesLongPollTimeout(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.pollDelay = 250 * time.Millisecond
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-1",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "hello",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-1",
	})
	plane, err := New(Config{
		BaseURL:         fake.URL(),
		DaemonID:        "daemon-1",
		Agents:          []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}},
		RequestTimeout:  50 * time.Millisecond,
		LongPollTimeout: 2 * time.Second,
		ClaimWaitMs:     20000,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	if req == nil || req.ID != "task-a" {
		t.Fatalf("claim = %+v", req)
	}
	if fake.lastPollWaitMs != 20000 {
		t.Fatalf("poll wait_ms = %d, want 20000", fake.lastPollWaitMs)
	}
}

// With ClaimWaitMs unset the daemon sends no long-poll hint (legacy short-poll),
// keeping the wire compatible with an old control plane.
func TestPlaneClaimOmitsWaitMsWhenDisabled(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-1",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "hello",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-1",
	})
	plane, err := New(Config{
		BaseURL:  fake.URL(),
		DaemonID: "daemon-1",
		Agents:   []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}},
		// ClaimWaitMs left zero.
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer plane.Close()

	if _, err := plane.ClaimTask(context.Background(), "daemon-1:codex"); err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	if fake.lastPollWaitMs != 0 {
		t.Fatalf("poll wait_ms = %d, want 0 (long-poll disabled)", fake.lastPollWaitMs)
	}
}
