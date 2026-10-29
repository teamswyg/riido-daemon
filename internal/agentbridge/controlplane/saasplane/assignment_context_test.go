package saasplane

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestFetchAssignmentContextSendsDeviceCredentialOnce(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var in AssignmentContextRequest
		_ = json.NewDecoder(r.Body).Decode(&in)
		if r.URL.Path != "/v1/agents/agent-1/assignment-context" || r.Header.Get("X-Riido-Device-Secret") != "device-secret" || in.LeaseToken != "lease-1" {
			t.Errorf("request path=%s lease=%q", r.URL.Path, in.LeaseToken)
		}
		if in.View == "agents" {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"schema_version":"v1","error":"assignment context read limit exceeded; start a new execution"}`))
			return
		}
		if in.View == "tree" {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"schema_version":"v1","error":"assignment context read record unavailable"}`))
			return
		}
		_, _ = w.Write([]byte(`{"assignment_id":"asg-1","view":"task","data":{"x":1},"sources":[{"kind":"task","id":"t","part":"body"}]}`))
	}))
	defer server.Close()
	plane, err := New(Config{BaseURL: server.URL, DaemonID: "daemon-1", DeviceID: "device-1", DeviceSecret: "device-secret", RequestTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	out, err := plane.FetchAssignmentContext(context.Background(), "agent-1", AssignmentContextRequest{AssignmentID: "asg-1", LeaseToken: "lease-1", View: "task"})
	if err != nil || string(out.Data) != `{"x":1}` || len(out.Sources) != 1 {
		t.Fatalf("out=%+v err=%v", out, err)
	}
	_, err = plane.FetchAssignmentContext(context.Background(), "agent-1", AssignmentContextRequest{AssignmentID: "asg-1", LeaseToken: "lease-1", View: "tree"})
	status, message, ok := HTTPStatus(err)
	if !ok || status != http.StatusServiceUnavailable || message != "assignment context read record unavailable" {
		t.Fatalf("status=%d message=%q ok=%v", status, message, ok)
	}
	_, err = plane.FetchAssignmentContext(context.Background(), "agent-1", AssignmentContextRequest{AssignmentID: "asg-1", LeaseToken: "lease-1", View: "agents"})
	status, _, ok = HTTPStatus(err)
	if !ok || status != http.StatusTooManyRequests {
		t.Fatalf("quota status lost: status=%d ok=%v", status, ok)
	}
	if calls.Load() != 3 {
		t.Fatalf("a recorded read must not be retried blindly, calls=%d", calls.Load())
	}
}
