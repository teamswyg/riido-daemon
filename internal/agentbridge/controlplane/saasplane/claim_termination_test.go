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
	"github.com/teamswyg/riido-daemon/internal/agentbridge/executionproof"
)

func TestTerminationDeliveryRetainsOriginalLease(t *testing.T) {
	var mu sync.Mutex
	events, receipts := 0, 0
	var keys []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if strings.HasSuffix(r.URL.Path, "/events") {
			events++
			var event assignmentcontract.AgentEventRequest
			if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
				t.Error(err)
			}
			keys = append(keys, event.Metadata["assignment_event_key"])
			if event.RuntimeID != "original-runtime" {
				t.Errorf("runtime changed: %q", event.RuntimeID)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{}`))
			return
		}
		if strings.HasSuffix(r.URL.Path, "/termination") {
			receipts++
			if r.Header.Get("X-Riido-Assignment-Lease-Token") != "fixture-lease" || r.Header.Get("X-Riido-Device-Secret") != "fixture-device-secret" || r.Header.Get("Authorization") != "" {
				t.Error("receipt authentication changed")
			}
			if receipts == 1 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			h := sha256.Sum256([]byte("fixture-lease"))
			_ = json.NewEncoder(w).Encode(map[string]any{"schema_version": "riido-execution-termination.v1", "kind": "local_invocations_quiescent", "assignment_id": "asn-proof", "lease_epoch": hex.EncodeToString(h[:]), "started_invocations": 0, "recorded_at": time.Now().UTC().Format(time.RFC3339Nano)})
			return
		}
		t.Errorf("unexpected path %s", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	p, err := New(Config{BaseURL: server.URL, DaemonID: "daemon-1", DeviceID: "device-1", DeviceSecret: "fixture-device-secret", BearerToken: "fixture-bearer"})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	a := assignmentLifecycleFixture("asn-proof", "prompt", "fixture-lease")
	req, err := p.claimStartOrActiveAssignment(context.Background(), "original-runtime", "codex", assignmentcontract.PollResponse{Action: assignmentcontract.PollStart, Assignment: &a})
	if err != nil {
		t.Fatal(err)
	}
	// The source must attach coverage; consumers cannot derive it from provider metadata.
	ctx := claimProofContext(t, req.LocalExecutionScope)
	if err = p.CompleteTask(ctx, req.ID, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "done"}); err == nil {
		t.Fatal("receipt failure must retain terminal delivery")
	}
	if _, ok, e := p.assignmentForExecution(context.Background(), req.ID); e != nil || !ok {
		t.Fatal("original lease cleared before receipt ACK")
	}
	if err = p.CompleteTask(ctx, req.ID, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "different retry input"}); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if events != 1 || receipts != 2 || keys[0] == "" {
		t.Fatalf("delivery/key mismatch: events=%d receipts=%d keys=%v", events, receipts, keys)
	}
}

func claimProofContext(t *testing.T, scope *executionproof.Scope) context.Context {
	t.Helper()
	if scope == nil {
		t.Fatal("source omitted private scope")
	}
	proof, ok := scope.Tracker().Seal()
	if !ok {
		t.Fatal("fresh idle claim missing coverage")
	}
	return executionproof.ContextWithProof(executionproof.ContextWithScope(context.Background(), scope), proof)
}
