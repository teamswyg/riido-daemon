package saasplane

import (
	"context"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestRecoveryMetadataIsBoundedAndEventKeyIsStable(t *testing.T) {
	assignment := saasClaimReportAssignment()
	c := assignmentcontract.RecoveryContext{State: "checking", Provider: "codex", Reason: "none", Circuit: "closed", OutcomeCertainty: "not_started", Attempt: 2}
	ev := agentbridge.Event{Kind: agentbridge.EventProgress, Text: "untrusted raw text", RecoveryContext: &c}
	request, ok := eventRequestFromAgentEvent(assignment, ev)
	if !ok || request.EventType != assignmentcontract.EventRiidoLog || request.State != "" || request.Message == ev.Text {
		t.Fatalf("recovery altered lifecycle or exposed raw message: %+v", request)
	}
	first := recoveryEventKey(assignment, "run", c)
	if first != recoveryEventKey(assignment, "run", c) {
		t.Fatal("retransmission changed idempotency key")
	}
	confirmed := c
	confirmed.OutcomeCertainty = "confirmed"
	if first == recoveryEventKey(assignment, "run", confirmed) {
		t.Fatal("started and confirmed observations collided")
	}
	for _, change := range []func(*assignmentcontract.RecoveryContext){func(c *assignmentcontract.RecoveryContext) { c.Reason = "transient_preflight" }, func(c *assignmentcontract.RecoveryContext) { c.Circuit = "half_open" }} {
		next := c
		change(&next)
		if first == recoveryEventKey(assignment, "run", next) {
			t.Fatal("meaningful recovery observation collided")
		}
	}
	c.Attempt++
	if first == recoveryEventKey(assignment, "run", c) {
		t.Fatal("new attempt reused idempotency key")
	}
	metadata := (assignmentcontract.RecoveryContext{State: "secret", Reason: "secret", Provider: "secret"}).Metadata()
	if metadata[assignmentcontract.RecoveryStateKey] != "unknown" || metadata[assignmentcontract.RecoveryReasonKey] != "unknown" || metadata[assignmentcontract.RecoveryProviderKey] != "unknown" {
		t.Fatal("unsafe diagnostic metadata escaped")
	}
}

func TestFallbackNativeRuntimeDoesNotRenewOriginalAssignmentLease(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()
	assignment := saasClaimReportAssignment()
	id := assignmentExecutionID(assignment)
	if err := plane.withState(context.Background(), func(s *planeState) {
		s.assignmentsByExecution[id] = assignment
		s.runtimeIDsByExecution[id] = "original-runtime"
	}); err != nil {
		t.Fatal(err)
	}
	wrong, err := plane.activeAssignmentsForRuntimeHeartbeat(context.Background(), "fallback-native-runtime", []string{id})
	if err != nil {
		t.Fatal(err)
	}
	if len(wrong) != 0 {
		t.Fatal("target native heartbeat acquired original runtime lease")
	}
	correct, err := plane.activeAssignmentsForRuntimeHeartbeat(context.Background(), "original-runtime", []string{id})
	if err != nil {
		t.Fatal(err)
	}
	if len(correct[assignment.AgentID]) != 1 {
		t.Fatal("origin heartbeat lost assignment ownership")
	}
}

func TestAssignmentRecoveryPolicyIsIndependentOfPollSnapshot(t *testing.T) {
	assignment := saasClaimReportAssignment()
	assignment.RecoveryPolicy = &assignmentcontract.ProviderRecoveryPolicy{MaxSwitches: 1, Fallbacks: []assignmentcontract.ProviderRecoveryCandidate{{Provider: "claude", ModelID: "sonnet"}}}
	request := taskRequestFromAssignment(assignment)
	assignment.RecoveryPolicy.Fallbacks[0].ModelID = "changed"
	if request.RecoveryPolicy.Fallbacks[0].ModelID != "sonnet" || request.ProviderNeutralPrompt != assignment.Prompt || request.AgentInstruction != assignment.AgentInstruction {
		t.Fatal("recovery input lost immutable authorization or neutral source")
	}
}

func TestRecoveryJournalIsSeparateFromExistingTerminalPayload(t *testing.T) {
	for _, status := range []agentbridge.ResultStatus{agentbridge.ResultCompleted, agentbridge.ResultFailed} {
		t.Run(string(status), func(t *testing.T) {
			fake := newFakeAssignmentServer(t)
			assignment := saasClaimReportAssignment()
			assignment.ResumeSessionID = ""
			fake.enqueue(assignment)
			plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
			defer plane.Close()
			request, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
			if err != nil {
				t.Fatal(err)
			}
			c := assignmentcontract.RecoveryContext{State: "recovered", Provider: "claude", Reason: "none", Circuit: "closed", Attempt: 2, SwitchCount: 1, OutcomeCertainty: "confirmed"}
			if err = plane.CompleteTask(context.Background(), request.ID, agentbridge.Result{Status: status, Output: "fixture result", SessionID: "foreign-native-session", RecoveryContext: &c}); err != nil {
				t.Fatal(err)
			}
			fake.mu.Lock()
			events := append([]assignmentcontract.AgentEventRequest(nil), fake.events...)
			fake.mu.Unlock()
			if len(events) != 2 || events[0].EventType != assignmentcontract.EventRiidoLog || events[0].State != "" {
				t.Fatalf("recovery changed assignment lifecycle: %+v", events)
			}
			terminal := events[1]
			if assignmentcontract.RecoveryContextFromMetadata(terminal.Metadata) != nil || terminal.ProviderSessionID != "" {
				t.Fatal("terminal crossed journal or native provider boundary")
			}
			if terminal.EventType != assignmentcontract.EventAssignmentCompleted && terminal.EventType != assignmentcontract.EventAssignmentFailed {
				t.Fatalf("missing ordinary terminal event: %s", terminal.EventType)
			}
		})
	}
}

func TestRecoveryJournalFailureDoesNotBlockTerminalDelivery(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	assignment := saasClaimReportAssignment()
	fake.enqueue(assignment)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()
	request, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatal(err)
	}
	fake.mu.Lock()
	fake.transientFailures["/v1/agents/jykim1/events"] = 1
	fake.transientStatuses["/v1/agents/jykim1/events"] = 403
	fake.mu.Unlock()
	c := assignmentcontract.RecoveryContext{State: "recovered", Provider: "codex", Reason: "none", Attempt: 2, OutcomeCertainty: "confirmed", Circuit: "closed"}
	if err = plane.CompleteTask(context.Background(), request.ID, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "done", RecoveryContext: &c}); err != nil {
		t.Fatalf("journal rejection prevented completion: %v", err)
	}
	fake.assertEvent(t, assignmentcontract.EventAssignmentCompleted)
}

func TestRecoveryRequiresActualAssignmentAcknowledgement(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()
	event := agentbridge.Event{Kind: agentbridge.EventProgress, RecoveryContext: &assignmentcontract.RecoveryContext{State: "checking", Provider: "codex"}}
	if err := plane.ReportEvent(context.Background(), "missing-execution", event); err == nil {
		t.Fatal("missing local authority was treated as a successful preflight acknowledgement")
	}
	if err := plane.ReportEvent(context.Background(), "missing-execution", agentbridge.Event{Kind: agentbridge.EventProgress}); err != nil {
		t.Fatal("ordinary legacy progress behavior changed")
	}
}
