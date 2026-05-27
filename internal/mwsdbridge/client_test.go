package mwsdbridge

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestFetchSnapshot(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "mwsd.sock")
	stop := serveFakeMwsd(t, socketPath, map[string]string{
		"status": `{
			"root": "/workspace",
			"socket_path": "/tmp/mwsd.sock",
			"graph_schema_version": "mws-doc-graph.v1",
			"domain_schema_version": "mws-cl-domain.v1",
			"harness_schema_version": "mws-harness-run.v1",
			"document_count": 23,
			"repository_count": 3,
			"domain_name": "macmini-workspace",
			"harness_run_count": 2,
			"harness_next_direction": "top-down",
			"harness_recent_directions": ["top-down", "bottom-up"],
			"ssot_conflict_count": 0,
			"domain_diagnostic_count": 0,
			"harness_diagnostic_count": 0,
			"orchestration_schema_version": "mws-orchestration-snapshot.v1"
		}`,
		"graph": `{
			"schema_version": "mws-doc-graph.v1",
			"root": "/workspace",
			"stats": {
				"document_count": 23,
				"node_count": 23,
				"edge_count": 100,
				"diagnostic_count": 0,
				"error_count": 0,
				"warning_count": 0,
				"unresolved_link_count": 0
			}
		}`,
		"domain": `{
			"schema_version": "mws-cl-domain.v1",
			"path": "/workspace/domains/macmini-workspace.lisp",
			"domain": "macmini-workspace",
			"repositories": [{
				"name": "riido-daemon",
				"owner": "kimjooyoon",
				"visibility": "private",
				"ssot_scope": "project-daemon",
				"local_path": "/Users/teddy/github/kimjooyoon/riido-daemon",
				"remote": "https://github.com/teamswyg/riido-daemon",
				"role": "project-ssot",
				"consumes": ["mws-doc-graph", "mws-cl-domain"]
			}],
			"diagnostics": []
		}`,
		"harness": `{
			"schema_version": "mws-harness-run.v1",
			"path": "/workspace/harness/runs.jsonl",
			"run_count": 2,
			"top_down_count": 1,
			"bottom_up_count": 1,
			"last_direction": "bottom-up",
			"next_direction": "top-down",
			"consecutive_direction_count": 1,
			"recent_directions": ["top-down", "bottom-up"],
			"diagnostics": []
		}`,
		"orchestration": `{
			"schema_version": "mws-orchestration-snapshot.v1",
			"root": "/workspace",
			"domain_path": "/workspace/domains/macmini-workspace.lisp",
			"harness_run_path": "/workspace/harness/runs.jsonl",
			"domain_schema_version": "mws-cl-domain.v1",
			"harness_schema_version": "mws-harness-run.v1",
			"mode": "orchestration-over-choreography",
			"decision_gate": "human-approval-required",
			"decision_by": ["human"],
			"decision_llms": ["codex"],
			"provider_candidates": [
				{"id": "codex", "source_workflow": "provider-selection", "available": true, "approval_required": true},
				{"id": "claude-code", "source_workflow": "provider-selection", "available": true, "approval_required": true},
				{"id": "cursor", "source_workflow": "provider-selection", "available": true, "approval_required": true}
			],
			"recommended_provider": "codex",
			"recommended_decision_llm": "codex",
			"next_action": {
				"direction": "top-down",
				"command_surface": "mwsd harness + riido task queue + mws-viewer cockpit",
				"reason": "lift the latest bottom-up evidence into the next SSOT plan",
				"requires_human_approval": true
			},
			"top_down_count": 1,
			"bottom_up_count": 1,
			"last_direction": "bottom-up",
			"balanced": true,
			"direction_bias": false,
			"workflows": [{
				"name": "provider-selection",
				"top_down": ["goal", "constraints"],
				"bottom_up": ["capability", "history"],
				"decision_by": ["human"],
				"decision_llm": ["codex"],
				"providers": ["codex", "claude-code", "cursor"],
				"loop_steps": ["propose", "choose", "assign", "verify", "record"]
			}],
			"recent_runs": [{
				"id": "run-1",
				"direction": "bottom-up",
				"source": "mwsd",
				"provider": "rust-binary",
				"command": "verify",
				"result": "passed"
			}],
			"diagnostics": []
		}`,
		"projects": `{
			"schema_version": "mws-project-registry.v1",
			"root": "/workspace",
			"domain_path": "/workspace/domains/macmini-workspace.lisp",
			"repository_count": 1,
			"repositories": [{
				"name": "riido-daemon",
				"owner": "kimjooyoon",
				"visibility": "private",
				"ssot_scope": "project-daemon",
				"local_path": "/Users/teddy/github/kimjooyoon/riido-daemon",
				"remote": "https://github.com/teamswyg/riido-daemon",
				"role": "project-ssot",
				"consumes": ["mws-doc-graph", "mws-cl-domain"],
				"local_present": true,
				"git_present": true,
				"remote_matches": true
			}],
			"diagnostics": []
		}`,
	})
	defer stop()

	snapshot, err := NewClient(socketPath).FetchSnapshot(context.Background())
	if err != nil {
		t.Fatalf("FetchSnapshot returned error: %v", err)
	}
	if snapshot.Status.Root != "/workspace" {
		t.Fatalf("unexpected root: %s", snapshot.Status.Root)
	}
	if snapshot.Graph.Stats.DocumentCount != 23 {
		t.Fatalf("unexpected document count: %d", snapshot.Graph.Stats.DocumentCount)
	}
	if snapshot.Domain.Domain != "macmini-workspace" {
		t.Fatalf("unexpected domain: %s", snapshot.Domain.Domain)
	}
	if snapshot.Harness.NextDirection != "top-down" {
		t.Fatalf("unexpected next direction: %s", snapshot.Harness.NextDirection)
	}
	if got := snapshot.Projects.Repositories[0].Name; got != "riido-daemon" {
		t.Fatalf("unexpected project repository: %s", got)
	}
	if snapshot.Orchestration.RecommendedProvider != "codex" {
		t.Fatalf("unexpected recommended provider: %s", snapshot.Orchestration.RecommendedProvider)
	}
	if len(snapshot.Orchestration.ProviderCandidates) != 3 {
		t.Fatalf("unexpected provider candidate count: %d", len(snapshot.Orchestration.ProviderCandidates))
	}
}

func TestRequestRejectsNotOK(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "mwsd.sock")
	stop := serveRawMwsd(t, socketPath, func(method string) string {
		return `{"ok":false,"method":"` + method + `","data":null,"error":"not ready"}`
	})
	defer stop()

	var status Status
	err := NewClient(socketPath).Request(context.Background(), "status", &status)
	if err == nil {
		t.Fatal("Request should fail when mwsd returns ok=false")
	}
}

func serveFakeMwsd(t *testing.T, socketPath string, data map[string]string) func() {
	t.Helper()
	return serveRawMwsd(t, socketPath, func(method string) string {
		payload, ok := data[method]
		if !ok {
			return `{"ok":false,"method":"` + method + `","data":null,"error":"unknown method"}`
		}
		return `{"ok":true,"method":"` + method + `","data":` + payload + `,"error":null}`
	})
}

func serveRawMwsd(t *testing.T, socketPath string, respond func(method string) string) func() {
	t.Helper()
	_ = os.Remove(socketPath)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen unix socket: %v", err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer conn.Close()
				var req request
				if err := json.NewDecoder(conn).Decode(&req); err != nil {
					return
				}
				_, _ = conn.Write([]byte(respond(req.Method)))
			}(conn)
		}
	}()
	return func() {
		_ = listener.Close()
		<-done
		_ = os.Remove(socketPath)
	}
}
