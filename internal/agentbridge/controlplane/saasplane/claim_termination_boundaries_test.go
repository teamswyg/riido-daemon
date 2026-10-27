package saasplane

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/executionproof"
)

type terminationFixture struct {
	plane      *Plane
	mu         sync.Mutex
	events     []assignmentcontract.AgentEventRequest
	receipts   []terminationRequest
	reply      func(http.ResponseWriter, *http.Request, terminationRequest)
	eventReply func(http.ResponseWriter, *http.Request) bool
}

func newTerminationFixture(t *testing.T) *terminationFixture {
	t.Helper()
	f := &terminationFixture{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if strings.HasSuffix(r.URL.Path, "/events") {
			var event assignmentcontract.AgentEventRequest
			if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
				t.Error(err)
			}
			f.events = append(f.events, event)
			if f.eventReply != nil && f.eventReply(w, r) {
				return
			}
			_, _ = w.Write([]byte(`{}`))
			return
		}
		if !strings.HasSuffix(r.URL.Path, "/termination") {
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var req terminationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
		}
		f.receipts = append(f.receipts, req)
		if f.reply != nil {
			f.reply(w, r, req)
			return
		}
		writeTerminationACK(w, r, req)
	}))
	p, err := New(Config{BaseURL: server.URL, DaemonID: "daemon-1", DeviceID: "device-1", DeviceSecret: "fixture-secret"})
	if err != nil {
		t.Fatal(err)
	}
	f.plane = p
	t.Cleanup(func() { p.Close(); server.Close() })
	return f
}

func writeTerminationACK(w http.ResponseWriter, r *http.Request, req terminationRequest) {
	hash := sha256.Sum256([]byte(r.Header.Get("X-Riido-Assignment-Lease-Token")))
	_ = json.NewEncoder(w).Encode(terminationReceipt{SchemaVersion: req.SchemaVersion, Kind: req.Kind, AssignmentID: "asn-proof", LeaseEpoch: hex.EncodeToString(hash[:]), StartedInvocations: req.StartedInvocations, RecordedAt: time.Now().UTC()})
}

func claimTerminationFixture(t *testing.T, p *Plane, action assignmentcontract.PollAction, lease, runtime string) *bridge.TaskRequest {
	t.Helper()
	a := assignmentLifecycleFixture("asn-proof", "prompt", lease)
	a.ResumeSessionID = "fixture-native-session"
	req, err := p.claimStartOrActiveAssignment(context.Background(), runtime, "codex", assignmentcontract.PollResponse{Action: action, Assignment: &a})
	if err != nil || req == nil {
		t.Fatalf("claim failed: %v", err)
	}
	return req
}

func TestClaimProofFreshStartVsRecoveredActive(t *testing.T) {
	for _, tc := range []struct {
		name   string
		action assignmentcontract.PollAction
		fresh  bool
	}{{"fresh", assignmentcontract.PollStart, true}, {"recovered", assignmentcontract.PollActive, false}} {
		t.Run(tc.name, func(t *testing.T) {
			f := newTerminationFixture(t)
			req := claimTerminationFixture(t, f.plane, tc.action, "fixture-lease", "runtime-1")
			again := claimTerminationFixture(t, f.plane, assignmentcontract.PollStart, "fixture-lease", "runtime-2")
			if req.LocalExecutionScope != again.LocalExecutionScope {
				t.Fatal("same epoch coverage was replaced")
			}
			if _, ok := req.LocalExecutionScope.Tracker().Seal(); ok != tc.fresh {
				t.Fatalf("coverage=%t want %t", ok, tc.fresh)
			}
		})
	}
}

func TestTerminationUnknownSkipsReceipt(t *testing.T) {
	f := newTerminationFixture(t)
	req := claimTerminationFixture(t, f.plane, assignmentcontract.PollActive, "fixture-lease", "runtime-1")
	ctx := executionproof.ContextWithScope(context.Background(), req.LocalExecutionScope)
	if err := f.plane.CompleteTask(ctx, req.ID, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "done"}); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := f.plane.assignmentForExecution(context.Background(), req.ID); err != nil || ok {
		t.Fatal("unknown proof blocked business cleanup")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.events) != 1 || len(f.receipts) != 0 {
		t.Fatal("unknown sent a receipt")
	}
}

func TestTerminationNeedsUserInputDoesNotCloseClaim(t *testing.T) {
	f := newTerminationFixture(t)
	req := claimTerminationFixture(t, f.plane, assignmentcontract.PollStart, "fixture-lease", "runtime-1")
	ctx := claimProofContext(t, req.LocalExecutionScope)
	res := agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "what should I do first?"}
	if err := f.plane.CompleteTask(ctx, req.ID, res); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	if len(f.events) != 1 || f.events[0].State != assignmentcontract.AssignmentRunning || len(f.receipts) != 0 {
		t.Fatalf("non-terminal result sent receipt: events=%+v receipts=%v", f.events, f.receipts)
	}
	f.mu.Unlock()
	resumed := claimTerminationFixture(t, f.plane, assignmentcontract.PollActive, "fixture-lease", "runtime-1")
	if resumed.LocalExecutionScope == req.LocalExecutionScope {
		t.Fatal("resumed same epoch reused sealed tracker")
	}
	attempt, err := resumed.LocalExecutionScope.Tracker().BeginInvocation()
	if err != nil {
		t.Fatal("unknown resumed scope must still execute", err)
	}
	attempt.Started()
	attempt.Exited(true)
	if _, ok := resumed.LocalExecutionScope.Tracker().Seal(); ok {
		t.Fatal("resumed scope invented earlier coverage")
	}
}

func TestTerminationOldEpochDoesNotClearNewClaim(t *testing.T) {
	f := newTerminationFixture(t)
	old := claimTerminationFixture(t, f.plane, assignmentcontract.PollStart, "old-fixture-lease", "old-runtime")
	ctx := claimProofContext(t, old.LocalExecutionScope)
	current := claimTerminationFixture(t, f.plane, assignmentcontract.PollStart, "new-fixture-lease", "new-runtime")
	watcher := make(chan error)
	if err := f.plane.withState(context.Background(), func(s *planeState) {
		s.partialBodies[current.ID].text = "new private body"
		s.cancelWatchers[current.ID] = watcher
		a := s.assignmentsByExecution[current.ID]
		a.ConfirmationReason = "runtime_approval"
		s.assignmentsByExecution[current.ID] = a
	}); err != nil {
		t.Fatal(err)
	}
	if err := f.plane.CompleteTask(ctx, old.ID, agentbridge.Result{Status: agentbridge.ResultCompleted}); err != nil {
		t.Fatal(err)
	}
	if err := f.plane.withState(context.Background(), func(s *planeState) {
		a := s.assignmentsByExecution[current.ID]
		if a.LeaseToken != "new-fixture-lease" || a.ConfirmationReason != "runtime_approval" || s.runtimeIDsByExecution[current.ID] != "new-runtime" || s.executionScopes[current.ID] != current.LocalExecutionScope || s.partialBodies[current.ID].text != "new private body" || s.cancelWatchers[current.ID] != watcher {
			t.Error("old completion changed current claim")
		}
	}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-watcher:
		t.Fatal("old completion closed new cancel watcher")
	default:
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.events) != 1 || f.events[0].RuntimeID != "old-runtime" || strings.Contains(f.events[0].Message, "new private body") {
		t.Fatalf("old completion borrowed new claim data: %+v", f.events)
	}
}

func TestTerminationRejectsMalformedACKWithoutCredentialReflection(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
	}{{"wrong epoch", `{"schema_version":"riido-execution-termination.v1","kind":"local_invocations_quiescent","assignment_id":"asn-proof","lease_epoch":"wrong","recorded_at":"2026-10-28T00:00:00Z"}`, 200}, {"overlong", strings.Repeat("fixture-secret", 500), 200}, {"error body", "fixture-secret fixture-lease", 503}, {"unknown field", `{"raw":"fixture-secret"}`, 200}, {"redirect", "fixture-secret", 302}} {
		t.Run(tc.name, func(t *testing.T) {
			f := newTerminationFixture(t)
			f.reply = func(w http.ResponseWriter, _ *http.Request, _ terminationRequest) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}
			req := claimTerminationFixture(t, f.plane, assignmentcontract.PollStart, "fixture-lease", "runtime-1")
			err := f.plane.CompleteTask(claimProofContext(t, req.LocalExecutionScope), req.ID, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "done"})
			if err == nil || strings.Contains(err.Error(), "fixture-secret") || strings.Contains(err.Error(), "fixture-lease") {
				t.Fatalf("unsafe acknowledgement error: %v", err)
			}
			if _, ok, e := f.plane.assignmentForExecution(context.Background(), req.ID); e != nil || !ok {
				t.Fatal("invalid ACK cleared claim")
			}
		})
	}
}

func TestTerminationBindingDoesNotKeepToolAuthorityAfterCleanup(t *testing.T) {
	f := newTerminationFixture(t)
	req := claimTerminationFixture(t, f.plane, assignmentcontract.PollStart, "fixture-lease", "runtime-1")
	ctx := claimProofContext(t, req.LocalExecutionScope)
	if err := f.plane.CompleteTask(ctx, req.ID, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "done"}); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := f.plane.assignmentForExecution(ctx, req.ID); err != nil || ok {
		t.Fatal("retained receipt identity extended general assignment authority")
	}
}

func TestTerminationUnsupportedServerKeepsBusinessCompletion(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusMethodNotAllowed} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			f := newTerminationFixture(t)
			f.reply = func(w http.ResponseWriter, _ *http.Request, _ terminationRequest) { w.WriteHeader(status) }
			req := claimTerminationFixture(t, f.plane, assignmentcontract.PollStart, "fixture-lease", "runtime-1")
			if err := f.plane.CompleteTask(claimProofContext(t, req.LocalExecutionScope), req.ID, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "done"}); err != nil {
				t.Fatalf("old CP unsupported route blocked terminal ACK: %v", err)
			}
			if _, ok, err := f.plane.assignmentForExecution(context.Background(), req.ID); err != nil || ok {
				t.Fatal("old CP claim retained after successful business terminal")
			}
			f.mu.Lock()
			defer f.mu.Unlock()
			if len(f.events) != 1 || len(f.receipts) != 1 {
				t.Fatal("unsupported route caused hidden retry")
			}
		})
	}
}

func TestTerminationLostTerminalACKUsesSamePayload(t *testing.T) {
	f := newTerminationFixture(t)
	f.eventReply = func(w http.ResponseWriter, _ *http.Request) bool {
		if len(f.events) == 1 {
			_, _ = w.Write([]byte(`{`))
			return true
		}
		return false
	}
	req := claimTerminationFixture(t, f.plane, assignmentcontract.PollStart, "fixture-lease", "runtime-1")
	ctx := claimProofContext(t, req.LocalExecutionScope)
	if err := f.plane.CompleteTask(ctx, req.ID, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "original result"}); err == nil {
		t.Fatal("lost terminal ACK accepted")
	}
	if err := f.plane.CompleteTask(ctx, req.ID, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "changed retry input"}); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.events) != 2 || len(f.receipts) != 1 || f.events[0].Message != "original result" || f.events[0].Message != f.events[1].Message || f.events[0].Metadata["assignment_event_key"] != f.events[1].Metadata["assignment_event_key"] {
		t.Fatal("terminal ACK retry changed frozen payload/key")
	}
}

func TestTerminationReceiptRetryDoesNotDependOnCurrentEpoch(t *testing.T) {
	f := newTerminationFixture(t)
	f.reply = func(w http.ResponseWriter, r *http.Request, req terminationRequest) {
		if r.Header.Get("X-Riido-Assignment-Lease-Token") != "old-fixture-lease" {
			t.Error("receipt borrowed new lease")
		}
		if len(f.receipts) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		writeTerminationACK(w, r, req)
	}
	old := claimTerminationFixture(t, f.plane, assignmentcontract.PollStart, "old-fixture-lease", "old-runtime")
	ctx := claimProofContext(t, old.LocalExecutionScope)
	if err := f.plane.CompleteTask(ctx, old.ID, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "done"}); err == nil {
		t.Fatal("receipt failure accepted")
	}
	current := claimTerminationFixture(t, f.plane, assignmentcontract.PollStart, "new-fixture-lease", "new-runtime")
	if err := f.plane.CompleteTask(ctx, old.ID, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "ignored retry"}); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	if len(f.events) != 1 || len(f.receipts) != 2 {
		t.Error("confirmed terminal ACK was retransmitted after reassignment")
	}
	f.mu.Unlock()
	if err := f.plane.withState(context.Background(), func(s *planeState) {
		if s.executionScopes[current.ID] != current.LocalExecutionScope || s.assignmentsByExecution[current.ID].LeaseToken != "new-fixture-lease" {
			t.Error("old receipt ACK deleted new claim")
		}
	}); err != nil {
		t.Fatal(err)
	}
}

func TestTerminationRejectsDuplicateACKKeys(t *testing.T) {
	for _, prefix := range []string{`"lease_epoch":"wrong",`, `"schema_version":"wrong",`, `"started_invocations":7,`} {
		t.Run(prefix, func(t *testing.T) {
			f := newTerminationFixture(t)
			f.reply = func(w http.ResponseWriter, r *http.Request, req terminationRequest) {
				recorder := httptest.NewRecorder()
				writeTerminationACK(recorder, r, req)
				body := recorder.Body.String()
				_, _ = w.Write([]byte("{" + prefix + body[1:]))
			}
			req := claimTerminationFixture(t, f.plane, assignmentcontract.PollStart, "fixture-lease", "runtime-1")
			if err := f.plane.CompleteTask(claimProofContext(t, req.LocalExecutionScope), req.ID, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "done"}); err == nil {
				t.Fatal("duplicate ACK key accepted")
			}
			if _, ok, err := f.plane.assignmentForExecution(context.Background(), req.ID); err != nil || !ok {
				t.Fatal("ambiguous ACK cleared claim")
			}
		})
	}
}

func TestTerminationRejectsMissingOrNullACKCount(t *testing.T) {
	for _, replacement := range []string{"", `"started_invocations":null,`} {
		t.Run(replacement, func(t *testing.T) {
			f := newTerminationFixture(t)
			f.reply = func(w http.ResponseWriter, r *http.Request, req terminationRequest) {
				recorder := httptest.NewRecorder()
				writeTerminationACK(recorder, r, req)
				body := strings.Replace(recorder.Body.String(), `"started_invocations":0,`, replacement, 1)
				_, _ = w.Write([]byte(body))
			}
			req := claimTerminationFixture(t, f.plane, assignmentcontract.PollStart, "fixture-lease", "runtime-1")
			if err := f.plane.CompleteTask(claimProofContext(t, req.LocalExecutionScope), req.ID, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "done"}); err == nil {
				t.Fatal("absent/null ACK count accepted as zero")
			}
		})
	}
}

func TestTerminationOldPartialBodyCannotCrossEpochBetweenReadAndWrite(t *testing.T) {
	f := newTerminationFixture(t)
	old := claimTerminationFixture(t, f.plane, assignmentcontract.PollStart, "old-fixture-lease", "old-runtime")
	ctx := executionproof.ContextWithScope(context.Background(), old.LocalExecutionScope)
	// Capture the assignment exactly as ReportEvent does before another poll
	// replaces the lease, then exercise the subsequent accumulation state step.
	assignment, ok, err := f.plane.assignmentForExecution(ctx, old.ID)
	if err != nil || !ok {
		t.Fatal("fixture old assignment unavailable")
	}
	current := claimTerminationFixture(t, f.plane, assignmentcontract.PollStart, "new-fixture-lease", "new-runtime")
	if err = f.plane.withState(context.Background(), func(s *planeState) { s.partialBodies[current.ID].text = "new text" }); err != nil {
		t.Fatal(err)
	}
	if err = f.plane.accumulatePartialBody(ctx, assignment, old.ID, "old text"); err != nil {
		t.Fatal(err)
	}
	if err = f.plane.withState(context.Background(), func(s *planeState) {
		if s.partialBodies[current.ID].text != "new text" || !s.partialBodies[current.ID].firstDeliveredAt.IsZero() {
			t.Error("old body mutated current epoch")
		}
	}); err != nil {
		t.Fatal(err)
	}
	if err = f.plane.ReportEvent(ctx, old.ID, agentbridge.Event{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning}); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.events) != 0 {
		t.Fatal("old scoped body/running event posted under new epoch")
	}
}
