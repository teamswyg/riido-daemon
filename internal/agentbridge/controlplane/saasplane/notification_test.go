package saasplane

import (
	"context"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestNotificationWaitingAndResume(t *testing.T) {
	for _, test := range []struct {
		event  agentbridge.Event
		reason assignmentcontract.ConfirmationReason
	}{
		{agentbridge.Event{Kind: agentbridge.EventToolApprovalNeeded}, assignmentcontract.ConfirmationApprovalRequired},
		{agentbridge.Event{Kind: agentbridge.EventToolApprovalNeeded, Tool: agentbridge.ToolRef{Name: "AskUserQuestion"}}, assignmentcontract.ConfirmationAnswerRequired},
		{agentbridge.Event{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateWaitingUserInput}, assignmentcontract.ConfirmationAnswerRequired},
		{agentbridge.Event{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateWaitingToolApproval}, assignmentcontract.ConfirmationApprovalRequired},
	} {
		req, ok := eventRequestFromAgentEvent(assignmentcontract.Assignment{ID: "a", TaskID: "t"}, test.event)
		if !ok || req.EventType != assignmentcontract.EventAssignmentWaitingForUser || req.ConfirmationReason != test.reason || req.State != assignmentcontract.AssignmentRunning {
			t.Fatalf("waiting: %+v", req)
		}
	}
	for _, event := range []agentbridge.Event{{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning}, {Kind: agentbridge.EventToolCallStarted}} {
		req, ok := eventRequestFromAgentEvent(assignmentcontract.Assignment{}, event)
		if !ok || req.EventType != assignmentcontract.EventAssignmentRunning || req.ConfirmationReason != "" {
			t.Fatalf("resume: %+v", req)
		}
	}
	req, _ := eventRequestFromAgentEvent(assignmentcontract.Assignment{}, agentbridge.Event{Kind: agentbridge.EventError, Err: "retrying"})
	if req.EventType != assignmentcontract.EventProviderError || req.State != "" {
		t.Fatalf("transient error promoted: %+v", req)
	}
}

func TestAutomaticallyApprovedToolDoesNotRequestHumanConfirmation(t *testing.T) {
	_, ok := eventRequestFromAgentEvent(assignmentcontract.Assignment{}, agentbridge.Event{Kind: agentbridge.EventToolApprovalNeeded, AutomaticallyApproved: true})
	if ok {
		t.Fatal("auto-approved tool requested human confirmation")
	}
}

func TestAnswerResumeOnTextDelta(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	assignment := assignmentcontract.Assignment{ID: "answer-run", TaskID: "task", AgentID: "jykim1", RuntimeProvider: "codex", State: assignmentcontract.AssignmentRunning, ConfirmationReason: assignmentcontract.ConfirmationAnswerRequired}
	fake.enqueue(assignment)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()
	ctx := context.Background()
	if err := plane.saveAssignmentRuntime(ctx, assignment, "daemon-1:codex"); err != nil {
		t.Fatal(err)
	}
	if err := plane.ReportEvent(ctx, assignment.ID, agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "resumed"}); err != nil {
		t.Fatal(err)
	}
	current, ok, err := plane.assignmentForExecution(ctx, assignment.ID)
	if err != nil || !ok || current.ConfirmationReason != "" {
		t.Fatalf("answer still waiting: %+v %v", current, err)
	}
	if len(fake.events) == 0 || fake.events[0].EventType != assignmentcontract.EventAssignmentRunning {
		t.Fatalf("resume not delivered: %+v", fake.events)
	}
}
