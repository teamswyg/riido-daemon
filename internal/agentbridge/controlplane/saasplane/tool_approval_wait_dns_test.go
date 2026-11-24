package saasplane

import (
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestWaitToolApprovalRetriesDNSFailureWithinExistingBudget(t *testing.T) {
	for _, tc := range []struct {
		name      string
		failed    int
		recovered bool
		wantCalls int
	}{
		{"same approval recovers after one failed poll", jsonRequestMaxAttempts, true, jsonRequestMaxAttempts + 1},
		{"persistent failure exhausts the wait budget", jsonRequestMaxAttempts * jsonRequestMaxAttempts, false, jsonRequestMaxAttempts * jsonRequestMaxAttempts},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := newFakeAssignmentServer(t)
			fake.toolApprovalStatus = assignmentcontract.ApprovalApproved
			plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
			t.Cleanup(plane.Close)
			plane.cfg.RequestTimeout = 2 * time.Second
			assignment := deadlineWaitAssignment()
			transport := &approvalDNSTransport{
				remainingFailures: tc.failed,
				expectedPath:      toolApprovalWaitPath(assignment, "approval-1"),
				next:              fake.server.Client().Transport,
				t:                 t,
			}
			plane.client = &http.Client{Transport: transport}

			resolution, err := plane.waitToolApproval(t.Context(), assignment, "approval-1")
			if tc.recovered {
				if err != nil || !resolution.Approved {
					t.Fatalf("explicit server approval was not returned: %+v, error = %v", resolution, err)
				}
			} else {
				var dnsErr *net.DNSError
				if !errors.As(err, &dnsErr) || !dnsErr.IsNotFound || resolution.Approved {
					t.Fatalf("persistent lookup failure must fail closed: %+v, error = %v", resolution, err)
				}
			}
			if transport.calls != tc.wantCalls {
				t.Fatalf("wait HTTP attempts = %d, want %d", transport.calls, tc.wantCalls)
			}
			wantServerReads := tc.wantCalls - tc.failed
			if len(fake.toolApprovals) != 0 || len(fake.toolApprovalWaits) != wantServerReads {
				t.Fatalf("created %d approvals, server waits = %d, want %d", len(fake.toolApprovals), len(fake.toolApprovalWaits), wantServerReads)
			}
		})
	}
}

type approvalDNSTransport struct {
	t                 *testing.T
	expectedPath      string
	firstBody         []byte
	remainingFailures int
	calls             int
	next              http.RoundTripper
}

func (transport *approvalDNSTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	_ = req.Body.Close()
	if err != nil {
		return nil, err
	}
	if req.Method != http.MethodPost || req.URL.Path != transport.expectedPath {
		transport.t.Errorf("unexpected approval request: %s %s", req.Method, req.URL.Path)
	}
	if transport.calls == 0 {
		transport.firstBody = body
	} else if !bytes.Equal(transport.firstBody, body) {
		transport.t.Error("retried approval changed the request body")
	}
	transport.calls++
	if transport.remainingFailures > 0 {
		transport.remainingFailures--
		return nil, &net.OpError{
			Op: "dial", Net: "tcp",
			Err: &net.DNSError{Err: "no such host", Name: req.URL.Hostname(), IsNotFound: true},
		}
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	return transport.next.RoundTrip(req)
}
