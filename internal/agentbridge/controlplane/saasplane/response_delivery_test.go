package saasplane

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestFirstAnswerSurvivesFailedConfirmationTransition(t *testing.T) {
	fake, plane, req := newTextDeltaFixture(t)
	ctx := context.Background()
	if err := plane.withState(ctx, func(s *planeState) {
		a := s.assignmentsByExecution[req.ID]
		a.ConfirmationReason = "tool_approval"
		s.assignmentsByExecution[req.ID] = a
	}); err != nil {
		t.Fatal(err)
	}
	fake.failNext("/v1/agents/jykim1/events", 1, http.StatusBadRequest)
	if err := plane.ReportEvent(ctx, req.ID, agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "Only answer"}); err == nil {
		t.Fatal("expected transition delivery failure")
	}
	if err := plane.CompleteTask(ctx, req.ID, agentbridge.Result{Status: agentbridge.ResultCompleted}); err != nil {
		t.Fatal(err)
	}
	last := fake.events[len(fake.events)-1]
	if last.State != assignmentcontract.AssignmentCompleted || last.Message != "Only answer" {
		t.Fatalf("lost answer: %+v", last)
	}
}

func TestAnswerCompletionRequiresTextOrVerifiedResult(t *testing.T) {
	for _, tc := range []struct {
		name, output, partial string
		verified              bool
		want                  assignmentcontract.AssignmentState
	}{
		{name: "empty", want: assignmentcontract.AssignmentFailed},
		{name: "whitespace", output: " \n\t", want: assignmentcontract.AssignmentFailed},
		{name: "answer only", output: "Two sentence summary.", want: assignmentcontract.AssignmentCompleted},
		{name: "whitespace final uses stream", output: " \n", partial: "Actual streamed answer", want: assignmentcontract.AssignmentCompleted},
		{name: "verified artifact", verified: true, want: assignmentcontract.AssignmentCompleted},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake, plane, req := newTextDeltaFixture(t)
			if tc.partial != "" {
				reportTextDeltas(t, plane, req.ID, tc.partial)
			}
			result := agentbridge.Result{Status: agentbridge.ResultCompleted, Output: tc.output}
			if tc.verified {
				result.ExecutionContext = &assignmentcontract.ExecutionContext{VerificationStatus: "passed"}
			}
			if err := plane.CompleteTask(context.Background(), req.ID, result); err != nil {
				t.Fatal(err)
			}
			last := fake.events[len(fake.events)-1]
			if last.State != tc.want {
				t.Fatalf("state=%s want=%s", last.State, tc.want)
			}
			if tc.want == assignmentcontract.AssignmentFailed && !strings.Contains(last.Message, "without an answer") {
				t.Fatalf("missing cause: %s", last.Message)
			}
			if tc.partial != "" && last.Message != tc.partial {
				t.Fatalf("fallback=%q", last.Message)
			}
			if err := plane.withState(context.Background(), func(s *planeState) {
				if _, ok := s.partialBodies[req.ID]; ok {
					t.Error("completed state retained")
				}
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFirstResponseTimingIsStableAcrossSnapshotsAndReclaims(t *testing.T) {
	_, plane, req := newTextDeltaFixture(t)
	ctx := context.Background()
	var claimed, first, delivered time.Time
	if err := plane.withState(ctx, func(s *planeState) { claimed = s.partialBodies[req.ID].claimedAt }); err != nil {
		t.Fatal(err)
	}
	reportTextDeltas(t, plane, req.ID, "hello")
	if err := plane.withState(ctx, func(s *planeState) {
		first = s.partialBodies[req.ID].firstTextAt
		delivered = s.partialBodies[req.ID].firstDeliveredAt
	}); err != nil {
		t.Fatal(err)
	}
	if first.IsZero() || delivered.Before(first) || first.Before(claimed) {
		t.Fatal("invalid first-response timing")
	}
	a, _, err := plane.assignmentForExecution(ctx, req.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = plane.saveAssignmentRuntime(ctx, a, "daemon-1:codex"); err != nil {
		t.Fatal(err)
	}
	reportTextDeltas(t, plane, req.ID, " world")
	if err = plane.withState(ctx, func(s *planeState) {
		st := s.partialBodies[req.ID]
		if st.claimedAt != claimed || st.firstTextAt != first || st.firstDeliveredAt != delivered {
			t.Error("timing reset by snapshot/reclaim")
		}
	}); err != nil {
		t.Fatal(err)
	}
}
