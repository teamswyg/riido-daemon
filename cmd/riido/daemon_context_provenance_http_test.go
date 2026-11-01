package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"unicode/utf8"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/saasplane"
	"github.com/teamswyg/riido-daemon/internal/contextbroker"
)

func TestContextSnapshotActualHTTP(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var req saasplane.AssignmentContextRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error("invalid request")
		}
		if r.URL.Path != "/v1/agents/agent-one/assignment-context" || r.Method != http.MethodPost ||
			r.Header.Get("X-Riido-Device-Secret") != "synthetic-device-secret" || req.AssignmentID != "assignment-one" || req.LeaseToken != "synthetic-lease" || req.View != "base" {
			t.Error("device/lease request contract changed")
		}
		if err := json.NewEncoder(w).Encode(snapshotProvenanceResponse()); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()
	plane, err := saasplane.New(saasplane.Config{BaseURL: server.URL, DaemonID: "daemon-one", DeviceID: "device-one", DeviceSecret: "synthetic-device-secret"})
	if err != nil {
		t.Fatal(err)
	}
	env := readSnapshotProvenance(t, saasContextFetcher{plane: plane}, "")
	raw, err := json.Marshal(env)
	if err != nil || env.Error != nil || !strings.Contains(string(raw), `"snapshot_id":"assignment-one"`) {
		t.Fatal("actual HTTP response provenance was not delivered")
	}
	if strings.Contains(string(raw), "synthetic-device-secret") || strings.Contains(string(raw), "synthetic-lease") {
		t.Fatal("credential leaked")
	}
	if calls.Load() != 1 {
		t.Fatal("read retried or provenance performed another request")
	}
}

func TestContextSnapshotLegacyBudgetKeepsIdentity(t *testing.T) {
	p := &snapshotProvenancePlane{response: snapshotProvenanceResponse()}
	p.response.Data = json.RawMessage(`{"component":{"id":"task-one"},"document":{"content":"` + strings.Repeat("본문", 100) + `"}}`)
	env := readSnapshotProvenance(t, saasContextFetcher{plane: p}, "1")
	raw, err := json.Marshal(env.Data)
	if err != nil || env.Error != nil || !strings.Contains(string(raw), `"snapshot_id":"assignment-one"`) || !strings.Contains(string(raw), `"captured_at":"2026-11-01T12:00:00.123456789Z"`) {
		t.Fatal("legacy budget truncated snapshot identity")
	}
}

func TestContextSnapshotBudgetIncludesProvenance(t *testing.T) {
	for _, scoped := range []bool{false, true} {
		p := &snapshotProvenancePlane{response: snapshotProvenanceResponse()}
		if !scoped {
			p.response.Data = json.RawMessage(`{"component":{"id":"task-one","title":"original"}}`)
		}
		budget := strconv.Itoa(utf8.RuneCount(p.response.Data))
		env := readSnapshotProvenance(t, saasContextFetcher{plane: p}, budget)
		if env.Error != nil {
			t.Fatal(env.Error.Code)
		}
		warned := false
		for _, warning := range env.Warnings {
			warned = warned || strings.HasPrefix(warning.Code, "BUDGET_")
		}
		if !warned {
			t.Fatal("provenance pushed the response over budget without a warning")
		}
	}
}

func TestContextSnapshotOtherViewsKeepExistingData(t *testing.T) {
	for _, view := range []string{"task", "tree", "resources", "comments", "agents", "agent", "meetings", "meeting", "content"} {
		t.Run(view, func(t *testing.T) {
			p := &snapshotProvenancePlane{response: snapshotProvenanceResponse()}
			p.response.View = view
			p.response.Live = view == "agents" || view == "agent"
			q := contextbroker.Query{AgentID: "agent-one", AssignmentID: "assignment-one", TaskID: "task-one", View: view, Target: "related-work"}
			got, err := (saasContextFetcher{plane: p}).FetchAssignmentContext(context.Background(), q)
			if err != nil || got.SnapshotID != "" || string(got.Data) != string(p.response.Data) {
				t.Fatal("non-base response acquired provenance or changed its shape")
			}
		})
	}
}
