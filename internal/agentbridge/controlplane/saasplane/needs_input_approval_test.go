package saasplane

import (
	"context"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCompleteTaskReportsNeedsInputForNaturalLanguageApproval(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-approval-text",
		TaskID:          "task-approval-text",
		AgentID:         "agent-claude",
		RuntimeProvider: "claude",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-approval-text",
	})
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "agent-claude", RuntimeProvider: "claude"}})
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:claude")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	result := agentbridge.Result{
		Status: agentbridge.ResultCompleted,
		Output: "파일 생성이나 명령 실행이 필요한 작업이에요. 진행해도 괜찮다면 댓글로 알려주세요.",
	}
	if err := plane.CompleteTask(context.Background(), req.ID, result); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}

	last := fake.events[len(fake.events)-1]
	assertNeedsInputEvent(t, last)
	if last.Message != result.Output {
		t.Fatalf("needs-input approval message = %q", last.Message)
	}
}
