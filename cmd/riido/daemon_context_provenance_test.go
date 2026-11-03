package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/saasplane"
	"github.com/teamswyg/riido-daemon/internal/contextbroker"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

type snapshotProvenancePlane struct {
	response saasplane.AssignmentContextResponse
	calls    int
}

func (p *snapshotProvenancePlane) FetchAssignmentContext(_ context.Context, _ string, _ saasplane.AssignmentContextRequest) (saasplane.AssignmentContextResponse, error) {
	p.calls++
	return p.response, nil
}

func snapshotProvenanceResponse() saasplane.AssignmentContextResponse {
	return saasplane.AssignmentContextResponse{
		AssignmentID: "assignment-one", TaskID: "task-one", View: "base",
		CapturedAt: time.Date(2026, 11, 1, 12, 0, 0, 123456789, time.UTC),
		Data:       json.RawMessage(`{"component":{"id":"task-one","title":"original"},"exact":9007199254740993,"read_scope":{"schema_version":"riido-read-scope.v1","resources":[]}}`),
	}
}

func readSnapshotProvenance(t *testing.T, fetcher contextbroker.Fetcher, budget string) contextcli.Envelope {
	t.Helper()
	b := contextbroker.New("", "")
	b.Fetcher = fetcher
	token, err := b.Registry.Issue(contextbroker.Binding{RunKey: "run-one", TaskID: "task-one", WorkspaceID: "workspace-one", AgentID: "agent-one", AssignmentID: "assignment-one", LeaseToken: "synthetic-lease"})
	if err != nil {
		t.Fatal(err)
	}
	return b.Handle(contextcli.ContextReadRequest{
		SchemaVersion: contextcli.ContextIPCSchemaVersion, RunCapability: token, Command: "context get", RequestID: "test-provenance", Args: map[string]string{"budget_chars": budget},
	}, func(string) (bool, bool) { return true, true })
}

func TestContextSnapshotProvenance(t *testing.T) {
	for _, budget := range []string{"", "1", "10000"} {
		t.Run("budget-"+budget, func(t *testing.T) {
			p := &snapshotProvenancePlane{response: snapshotProvenanceResponse()}
			env := readSnapshotProvenance(t, saasContextFetcher{plane: p}, budget)
			if env.Error != nil {
				t.Fatal(env.Error.Code)
			}
			raw, err := json.Marshal(env.Data)
			if err != nil {
				t.Fatal(err)
			}
			var data map[string]json.RawMessage
			if err = json.Unmarshal(raw, &data); err != nil {
				t.Fatal(err)
			}
			if string(data["snapshot_id"]) != `"assignment-one"` || string(data["captured_at"]) != `"2026-11-01T12:00:00.123456789Z"` {
				t.Fatal("context get must expose verified snapshot identity and capture time")
			}
			if string(data["exact"]) != "9007199254740993" {
				t.Fatal("context number was rounded")
			}
			if p.calls != 1 {
				t.Fatal("provenance must not add another API request")
			}
		})
	}
}

func TestContextSnapshotRejectsMismatchedProvenance(t *testing.T) {
	for name, change := range map[string]func(*saasplane.AssignmentContextResponse){
		"assignment":         func(r *saasplane.AssignmentContextResponse) { r.AssignmentID = "other" },
		"missing-assignment": func(r *saasplane.AssignmentContextResponse) { r.AssignmentID = "" },
		"task":               func(r *saasplane.AssignmentContextResponse) { r.TaskID = "other" },
		"missing-task":       func(r *saasplane.AssignmentContextResponse) { r.TaskID = "" },
		"view":               func(r *saasplane.AssignmentContextResponse) { r.View = "agents" },
		"missing-view":       func(r *saasplane.AssignmentContextResponse) { r.View = "" },
		"live-base":          func(r *saasplane.AssignmentContextResponse) { r.Live = true },
		"no-time":            func(r *saasplane.AssignmentContextResponse) { r.CapturedAt = time.Time{} },
		"null-data":          func(r *saasplane.AssignmentContextResponse) { r.Data = json.RawMessage("null") },
		"array-data":         func(r *saasplane.AssignmentContextResponse) { r.Data = json.RawMessage("[]") },
	} {
		t.Run(name, func(t *testing.T) {
			p := &snapshotProvenancePlane{response: snapshotProvenanceResponse()}
			change(&p.response)
			env := readSnapshotProvenance(t, saasContextFetcher{plane: p}, "")
			if env.Error == nil || env.Error.Code != contextcli.CodeUnavailable || env.Data != nil {
				t.Fatal("unverified provenance must fail without data")
			}
		})
	}
}
