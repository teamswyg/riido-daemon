package saasplane

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestServerTerminalConflictStopsRepeatedDelivery(t *testing.T) {
	for _, state := range []agentbridge.ResultStatus{agentbridge.ResultFailed, agentbridge.ResultCancelled} {
		t.Run(string(state), func(t *testing.T) {
			var calls atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				var event assignmentcontract.AgentEventRequest
				if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
					t.Error(err)
				}
				if event.AssignmentID != "asn-terminal" || event.State != assignmentcontract.AssignmentState(state) {
					t.Errorf("unexpected request for %s", event.AssignmentID)
				}
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"schema_version": assignmentcontract.SchemaVersion, "error": "invalid assignment transition completed -> " + string(event.State)})
			}))
			defer srv.Close()
			p, err := New(Config{BaseURL: srv.URL, DaemonID: "daemon", DeviceID: "device", DeviceSecret: "fixture-secret"})
			if err != nil {
				t.Fatal(err)
			}
			defer p.Close()
			a := assignmentLifecycleFixture("asn-terminal", "fixture", "lease")
			req, err := p.claimStartOrActiveAssignment(context.Background(), "runtime", "codex", assignmentcontract.PollResponse{Action: assignmentcontract.PollStart, Assignment: &a})
			if err != nil {
				t.Fatal(err)
			}
			ctx := claimProofContext(t, req.LocalExecutionScope)
			for i := 0; i < 4; i++ {
				if err := p.CompleteTask(ctx, req.ID, agentbridge.Result{Status: state, Error: "fixture"}); err != nil {
					t.Fatal(err)
				}
			}
			if calls.Load() != 1 {
				t.Fatalf("expected one event and no false termination receipt; got %d calls", calls.Load())
			}
			if _, ok, err := p.assignmentForExecution(ctx, req.ID); err != nil || ok {
				t.Fatalf("retired assignment remains active: %t %v", ok, err)
			}
		})
	}
}

func TestOnlyAuthoritativeTerminalTransitionRetiresReport(t *testing.T) {
	for _, tc := range []struct {
		code         int
		version, msg string
		want         bool
	}{
		{400, assignmentcontract.SchemaVersion, "invalid assignment transition cancelled -> failed", true},
		{400, assignmentcontract.SchemaVersion, "invalid assignment transition ready -> failed", false},
		{400, assignmentcontract.SchemaVersion, "invalid assignment transition completed -> running", false},
		{400, assignmentcontract.SchemaVersion, "unknown assignment", false},
		{400, "", "invalid assignment transition completed -> failed", false},
		{503, assignmentcontract.SchemaVersion, "invalid assignment transition completed -> failed", false},
	} {
		b, _ := json.Marshal(map[string]string{"schema_version": tc.version, "error": tc.msg})
		if _, ok := confirmedTerminalConflict(httpStatusError{StatusCode: tc.code, Body: string(b)}, "failed"); ok != tc.want {
			t.Errorf("status %d %q = %v, want %v", tc.code, tc.msg, ok, tc.want)
		}
	}
}
