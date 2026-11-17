package saasplane

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestWaitToolApprovalRetriesChildRequestDeadline(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()
	transport := &approvalDeadlineTransport{deadlines: map[int]bool{1: true}, next: fake.server.Client().Transport}
	plane.client = &http.Client{Transport: transport}
	// Scale only this synthetic request deadline; production long-poll defaults stay unchanged.
	plane.cfg.RequestTimeout = 20 * time.Millisecond
	ctx := t.Context()
	assignment := assignmentcontract.Assignment{ID: "asn-1", AgentID: "jykim1"}

	resolution, err := plane.waitToolApproval(ctx, assignment, "approval-1")
	if ctx.Err() != nil {
		t.Fatalf("parent context ended: %v", ctx.Err())
	}
	if err != nil {
		t.Fatalf("child request deadline ended current approval wait: %v", err)
	}
	if !resolution.Approved {
		t.Fatalf("explicit server approval was not returned: %+v", resolution)
	}
	if len(transport.bodies) != 2 || transport.bodies[0] != transport.bodies[1] {
		t.Fatalf("wait bodies = %q, want two identical requests", transport.bodies)
	}
	wantPath := toolApprovalWaitPath(assignment, "approval-1")
	for _, path := range transport.paths {
		if path != wantPath {
			t.Fatalf("request path = %q, want same approval %q", path, wantPath)
		}
	}
	if len(fake.toolApprovals) != 0 || len(fake.toolApprovalWaits) != 1 {
		t.Fatalf("created %d approvals, completed %d waits", len(fake.toolApprovals), len(fake.toolApprovalWaits))
	}
}

type approvalDeadlineTransport struct {
	deadlines      map[int]bool
	beforeWait     func()
	beforeResponse func(int)
	next           http.RoundTripper
	bodies         []string
	paths          []string
}

func (t *approvalDeadlineTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	t.bodies = append(t.bodies, string(body))
	t.paths = append(t.paths, req.URL.Path)
	req.Body = io.NopCloser(bytes.NewReader(body))
	if t.deadlines[len(t.bodies)] {
		if t.beforeWait != nil {
			t.beforeWait()
		}
		<-req.Context().Done()
		return nil, req.Context().Err()
	}
	if t.beforeResponse != nil {
		t.beforeResponse(len(t.bodies))
	}
	return t.next.RoundTrip(req)
}
