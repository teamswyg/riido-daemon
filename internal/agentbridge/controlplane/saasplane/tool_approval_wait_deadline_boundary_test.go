package saasplane

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestWaitToolApprovalStopsAfterConsecutiveRequestDeadlines(t *testing.T) {
	plane, _, transport := newDeadlineWaitPlane(t, map[int]bool{1: true, 2: true, 3: true, 4: true})
	resolution, err := plane.waitToolApproval(context.Background(), deadlineWaitAssignment(), "approval-1")
	if !errors.Is(err, context.DeadlineExceeded) || resolution.Approved {
		t.Fatalf("resolution = %+v, error = %v", resolution, err)
	}
	if len(transport.bodies) != jsonRequestMaxAttempts {
		t.Fatalf("wait calls = %d, want %d", len(transport.bodies), jsonRequestMaxAttempts)
	}
}

func TestWaitToolApprovalDoesNotRetryParentDeadline(t *testing.T) {
	plane, _, transport := newDeadlineWaitPlane(t, map[int]bool{1: true})
	plane.cfg.RequestTimeout = time.Second
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	resolution, err := plane.waitToolApproval(ctx, deadlineWaitAssignment(), "approval-1")
	if !errors.Is(err, context.DeadlineExceeded) || ctx.Err() == nil || resolution.Approved {
		t.Fatalf("resolution = %+v, error = %v, parent = %v", resolution, err, ctx.Err())
	}
	if len(transport.bodies) != 1 {
		t.Fatalf("wait calls = %d, want 1", len(transport.bodies))
	}
}

func TestWaitToolApprovalResetsDeadlineCountAfterPending(t *testing.T) {
	plane, fake, transport := newDeadlineWaitPlane(t, map[int]bool{1: true, 3: true, 5: true})
	transport.beforeResponse = func(call int) {
		fake.toolApprovalStatus = assignmentcontract.ApprovalPending
		if call == 6 {
			fake.toolApprovalStatus = assignmentcontract.ApprovalApproved
		}
	}
	resolution, err := plane.waitToolApproval(context.Background(), deadlineWaitAssignment(), "approval-1")
	if err != nil || !resolution.Approved || len(transport.bodies) != 6 {
		t.Fatalf("resolution = %+v, error = %v, calls = %d", resolution, err, len(transport.bodies))
	}
}

func TestWaitToolApprovalDoesNotRetryParentCancellation(t *testing.T) {
	plane, _, transport := newDeadlineWaitPlane(t, map[int]bool{1: true})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	transport.beforeWait = cancel
	resolution, err := plane.waitToolApproval(ctx, deadlineWaitAssignment(), "approval-1")
	if !errors.Is(err, context.Canceled) || resolution.Approved || len(transport.bodies) != 1 {
		t.Fatalf("resolution = %+v, error = %v, calls = %d", resolution, err, len(transport.bodies))
	}
}

func TestWaitToolApprovalPreservesTerminalResponsesAfterDeadline(t *testing.T) {
	for _, status := range []assignmentcontract.ApprovalStatus{assignmentcontract.ApprovalDenied, assignmentcontract.ApprovalTimedOut} {
		t.Run(string(status), func(t *testing.T) {
			plane, fake, transport := newDeadlineWaitPlane(t, map[int]bool{1: true})
			fake.toolApprovalStatus = status
			resolution, err := plane.waitToolApproval(context.Background(), deadlineWaitAssignment(), "approval-1")
			if err != nil || resolution.Approved || len(transport.bodies) != 2 {
				t.Fatalf("resolution = %+v, error = %v, calls = %d", resolution, err, len(transport.bodies))
			}
		})
	}
}

func TestWaitToolApprovalDoesNotRecoverAuthorizationFailure(t *testing.T) {
	plane, fake, transport := newDeadlineWaitPlane(t, map[int]bool{1: true})
	fake.failNext(toolApprovalWaitPath(deadlineWaitAssignment(), "approval-1"), 1, http.StatusUnauthorized)
	resolution, err := plane.waitToolApproval(context.Background(), deadlineWaitAssignment(), "approval-1")
	var statusErr httpStatusError
	if !errors.As(err, &statusErr) || statusErr.StatusCode != http.StatusUnauthorized || resolution.Approved {
		t.Fatalf("resolution = %+v, error = %v", resolution, err)
	}
	if len(transport.bodies) != 2 {
		t.Fatalf("wait calls = %d, want 2", len(transport.bodies))
	}
}

func newDeadlineWaitPlane(t *testing.T, deadlines map[int]bool) (*Plane, *fakeAssignmentServer, *approvalDeadlineTransport) {
	t.Helper()
	fake := newFakeAssignmentServer(t)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	t.Cleanup(plane.Close)
	transport := &approvalDeadlineTransport{deadlines: deadlines, next: fake.server.Client().Transport}
	plane.client = &http.Client{Transport: transport}
	plane.cfg.RequestTimeout = 20 * time.Millisecond
	return plane, fake, transport
}

func deadlineWaitAssignment() assignmentcontract.Assignment {
	return assignmentcontract.Assignment{ID: "asn-1", AgentID: "jykim1"}
}
