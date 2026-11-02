package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/saasplane"
	"github.com/teamswyg/riido-daemon/internal/contextbroker"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func TestContentCacheActualHTTP(t *testing.T) {
	for _, scenario := range []string{"hit", "legacy", "changed_source", "bad_identity", "refused", "unsolicited"} {
		t.Run(scenario, func(t *testing.T) {
			calls := 0
			digest := "sha256:" + strings.Repeat("a", 64)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				raw, _ := io.ReadAll(r.Body)
				var body map[string]json.RawMessage
				if err := json.Unmarshal(raw, &body); err != nil {
					t.Error(err)
				}
				if r.Method != http.MethodPost || r.URL.Path != "/v1/agents/agent-one/assignment-context" || r.Header.Get("X-Riido-Device-Secret") != "synthetic-secret" {
					t.Error("wrong request auth/path")
				}
				if r.Header.Get("X-Riido-Context-Cache-Version") != "1" {
					t.Error("cache version not negotiated")
				}
				if _, ok := body["cache_enabled"]; ok {
					t.Error("new JSON breaks legacy decoder")
				}
				if _, ok := body["known_digest"]; ok {
					t.Error("digest must use header")
				}
				if calls == 1 && r.Header.Get("X-Riido-Context-Known-Digest") != "" {
					t.Error("cold read has digest")
				}
				if calls == 2 && scenario != "legacy" && scenario != "unsolicited" && r.Header.Get("X-Riido-Context-Known-Digest") != digest {
					t.Error("candidate digest missing")
				}
				if calls == 2 && scenario == "refused" {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				out := saasplane.AssignmentContextResponse{AssignmentID: "assignment-one", TaskID: "task-one", View: "content", CapturedAt: time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC), ContentDigest: digest, Data: json.RawMessage(`{"source_id":"source-one","cursor":"page-one","content":"body","next_cursor":"page-two"}`), Sources: []saasplane.AssignmentContextSource{{Kind: "task", ID: "source-one", Part: "body", Version: "v1", UpdatedAt: "2026-11-01T00:00:00Z"}}}
				if scenario == "legacy" {
					out.ContentDigest = ""
				} else if calls == 2 || scenario == "unsolicited" {
					out.NotModified = true
					out.Data = json.RawMessage("null")
				}
				if calls == 2 && scenario == "changed_source" {
					out.Sources[0].UpdatedAt = "2026-11-02T00:00:00Z"
				}
				if calls == 2 && scenario == "bad_identity" {
					out.AssignmentID = "different-assignment"
				}
				if err := json.NewEncoder(w).Encode(out); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			plane, err := saasplane.New(saasplane.Config{BaseURL: server.URL, DaemonID: "daemon-one", DeviceID: "device-one", DeviceSecret: "synthetic-secret"})
			if err != nil {
				t.Fatal(err)
			}
			defer plane.Close()
			broker := contextbroker.New("", "")
			broker.Fetcher = saasContextFetcher{plane: plane}
			token, err := broker.Registry.Issue(contextbroker.Binding{RunKey: "run-one", AgentID: "agent-one", AssignmentID: "assignment-one", TaskID: "task-one", LeaseToken: "synthetic-lease"})
			if err != nil {
				t.Fatal(err)
			}
			req := contextcli.ContextReadRequest{SchemaVersion: contextcli.ContextIPCSchemaVersion, RunCapability: token, Command: "context part", RequestID: "read-1", Args: map[string]string{"work": "source-one", "cursor": "page-one"}}
			active := func(string) (bool, bool) { return true, true }
			first := broker.Handle(req, active)
			if scenario == "unsolicited" {
				if first.Error == nil || first.Data != nil {
					t.Fatal("unsolicited reuse accepted")
				}
				return
			}
			if first.Error != nil {
				t.Fatal(first.Error)
			}
			req.RequestID = "read-2"
			second := broker.Handle(req, active)
			if calls != 2 {
				t.Fatal("cache bypassed HTTP or retried")
			}
			if scenario == "hit" || scenario == "legacy" {
				if second.Error != nil {
					t.Fatal(second.Error)
				}
				a, _ := json.Marshal(first.Data)
				b, _ := json.Marshal(second.Data)
				if string(a) != string(b) || second.Page == nil || second.Page.NextCursor != "page-two" {
					t.Fatal("content/cursor changed")
				}
			} else if second.Error == nil || second.Data != nil || len(second.Sources) != 0 {
				t.Fatal("invalid response returned data")
			}
			encoded, _ := json.Marshal(second)
			for _, secret := range []string{digest, "synthetic-secret", "synthetic-lease", "not_modified", "content_digest"} {
				if strings.Contains(string(encoded), secret) {
					t.Fatal("internal cache/auth metadata leaked")
				}
			}
		})
	}
}
