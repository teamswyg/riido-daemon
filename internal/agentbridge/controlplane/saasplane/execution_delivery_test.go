package saasplane

import (
	"context"
	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"strings"
	"testing"
	"time"
)

func TestWaitingApprovalIsVisible(t *testing.T) {
	_, ok := eventRequestFromAgentEvent(assignmentcontract.Assignment{ID: "a", TaskID: "t"}, agentbridge.Event{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateWaitingToolApproval})
	if !ok {
		t.Fatal("waiting_tool_approval is dropped before control-plane")
	}
}
func TestWarningKeepsCause(t *testing.T) {
	req, ok := eventRequestFromAgentEvent(assignmentcontract.Assignment{ID: "a", TaskID: "t"}, agentbridge.Event{Kind: agentbridge.EventWarning, Text: "claude rate limited", Err: "retry after 120 seconds"})
	if !ok || !strings.Contains(req.Message, "120") {
		t.Fatalf("cause lost: %+v", req)
	}
}
func TestPartialFlushesWithoutNextDelta(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.enqueue(assignmentcontract.Assignment{ID: "a", TaskID: "t", ComponentID: "c", AgentID: "jykim1", RuntimeProvider: "codex", Prompt: "test", State: assignmentcontract.AssignmentQueued, LeaseToken: "l"})
	p := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer p.Close()
	ctx := context.Background()
	req, e := p.ClaimTask(ctx, "daemon-1:codex")
	if e != nil {
		t.Fatal(e)
	}
	for _, s := range []string{"a", "b"} {
		if e := p.ReportEvent(ctx, req.ID, agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: s}); e != nil {
			t.Fatal(e)
		}
	}
	time.Sleep(450 * time.Millisecond)
	if got := fake.events[len(fake.events)-1].Message; got != "ab" {
		t.Fatalf("after 450ms without next delta, visible=%q, want ab", got)
	}
}
func TestSameTaskAssignmentsStayIsolated(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	agents := []AgentBinding{{AgentID: "agent-a", RuntimeProvider: "claude"}, {AgentID: "agent-b", RuntimeProvider: "codex"}}
	for i, a := range agents {
		id := []string{"assignment-a", "assignment-b"}[i]
		fake.enqueue(assignmentcontract.Assignment{ID: id, TaskID: "shared-task", ComponentID: "c", AgentID: a.AgentID, RuntimeProvider: a.RuntimeProvider, Prompt: "test", State: assignmentcontract.AssignmentQueued, LeaseToken: id})
	}
	p := newTestPlane(t, fake.URL(), agents)
	defer p.Close()
	ctx := context.Background()
	a, e := p.ClaimTask(ctx, RuntimeIDForAgent("daemon-1", agents[0]))
	if e != nil || a == nil {
		t.Fatalf("claim a: %v", e)
	}
	b, e := p.ClaimTask(ctx, RuntimeIDForAgent("daemon-1", agents[1]))
	if e != nil || b == nil {
		t.Fatalf("claim b: %v", e)
	}
	if e := p.ReportEvent(ctx, a.ID, agentbridge.Event{Kind: agentbridge.EventProgress, Text: "agent-a progress"}); e != nil {
		t.Fatal(e)
	}
	got := fake.events[len(fake.events)-1]
	if got.AssignmentID != "assignment-a" {
		t.Fatalf("agent-a progress misrouted to %s", got.AssignmentID)
	}
}

func TestHeartbeatDeliversCancellationForRefreshedAssignment(t *testing.T) {
	plane, err := New(Config{BaseURL: "http://localhost", DaemonID: "daemon", BearerToken: "test", Agents: []AgentBinding{{AgentID: "agent", RuntimeProvider: "codex"}}})
	if err != nil {
		t.Fatal(err)
	}
	defer plane.Close()
	ctx := context.Background()
	assignment := assignmentcontract.Assignment{ID: "assignment-cancel", TaskID: "task", AgentID: "agent", State: assignmentcontract.AssignmentRunning}
	if err := plane.withState(ctx, func(s *planeState) { s.assignmentsByRun[assignment.ID] = assignment }); err != nil {
		t.Fatal(err)
	}
	cancelled, err := plane.WatchCancellation(ctx, assignment.ID)
	if err != nil {
		t.Fatal(err)
	}
	assignment.State = assignmentcontract.AssignmentCancelling
	if err := plane.deliverUnrefreshedHeartbeatCancels(ctx, []string{assignment.ID}, assignmentcontract.AgentHeartbeatResponse{RefreshedAssignments: []assignmentcontract.Assignment{assignment}}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-cancelled:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("heartbeat cancellation ignored while claim loop waits for runtime capacity")
	}
	if _, ok, err := plane.assignmentForTask(ctx, assignment.ID); err != nil || !ok {
		t.Fatal("cancellation lost identity required to report terminal result")
	}
}
