package main

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/mwsdbridge"
	"github.com/teamswyg/riido-daemon/internal/project"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestMwsdSyncWritesStateAndTaskDB(t *testing.T) {
	socketPath, stop := serveTestMwsd(t)
	defer stop()
	statePath := filepath.Join(t.TempDir(), "workspace-state.json")
	taskDBPath := filepath.Join(t.TempDir(), "task-db.json")

	out := captureStdout(t, func() {
		if err := run([]string{
			"mwsd", "sync",
			"--socket", socketPath,
			"--state", statePath,
			"--task-db", taskDBPath,
		}); err != nil {
			t.Fatalf("run mwsd sync: %v", err)
		}
	})
	if !json.Valid([]byte(out)) {
		t.Fatalf("sync output is not JSON: %s", out)
	}
	state, err := project.LoadState(statePath)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if len(state.Tasks) != 1 || state.Tasks[0].ID != "task:mws.cli" {
		t.Fatalf("unexpected state tasks: %#v", state.Tasks)
	}
	db, err := taskdb.LoadTaskDB(taskDBPath)
	if err != nil {
		t.Fatalf("LoadTaskDB: %v", err)
	}
	if len(db.Tasks) != 1 || db.Tasks[0].ID != "task:mws.cli" {
		t.Fatalf("unexpected task DB tasks: %#v", db.Tasks)
	}
	if len(db.Transitions) != 1 {
		t.Fatalf("expected one created transition, got %d", len(db.Transitions))
	}
}

func TestMwsdProjectionPrintsWorkspaceProjection(t *testing.T) {
	socketPath, stop := serveTestMwsd(t)
	defer stop()

	out := captureStdout(t, func() {
		if err := run([]string{"mwsd", "projection", "--socket", socketPath}); err != nil {
			t.Fatalf("run mwsd projection: %v", err)
		}
	})
	var projection project.WorkspaceProjection
	if err := json.Unmarshal([]byte(out), &projection); err != nil {
		t.Fatalf("parse projection output: %v\n%s", err, out)
	}
	if projection.SchemaVersion != "riido-workspace-projection.v1" {
		t.Fatalf("unexpected projection schema: %s", projection.SchemaVersion)
	}
	if len(projection.DocumentTaskLinks) != 1 || projection.DocumentTaskLinks[0].TaskID != "task:mws.cli" {
		t.Fatalf("unexpected projection task links: %#v", projection.DocumentTaskLinks)
	}
}

func serveTestMwsd(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("/tmp", "riido-mwsd-test-")
	if err != nil {
		t.Fatalf("create short socket dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	socketPath := filepath.Join(dir, "mwsd.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen unix: %v", err)
	}
	done := make(chan struct{})
	snapshot := cliMwsdSnapshot()
	go func() {
		defer close(done)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go serveTestMwsdConn(conn, snapshot)
		}
	}()
	return socketPath, func() {
		_ = listener.Close()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for mwsd test server shutdown")
		}
	}
}

func serveTestMwsdConn(conn net.Conn, snapshot mwsdbridge.Snapshot) {
	defer conn.Close()
	var req struct {
		Method string `json:"method"`
	}
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		return
	}
	var data any
	switch req.Method {
	case "status":
		data = snapshot.Status
	case "graph":
		data = snapshot.Graph
	case "domain":
		data = snapshot.Domain
	case "harness":
		data = snapshot.Harness
	case "orchestration":
		data = snapshot.Orchestration
	case "projects":
		data = snapshot.Projects
	default:
		_ = json.NewEncoder(conn).Encode(map[string]any{"ok": false, "method": req.Method, "error": "unknown method"})
		return
	}
	body, _ := json.Marshal(data)
	_ = json.NewEncoder(conn).Encode(struct {
		OK     bool            `json:"ok"`
		Method string          `json:"method"`
		Data   json.RawMessage `json:"data"`
	}{
		OK:     true,
		Method: req.Method,
		Data:   body,
	})
}

func cliMwsdSnapshot() mwsdbridge.Snapshot {
	root := "/tmp/riido-cli-mwsd"
	return mwsdbridge.Snapshot{
		Status: mwsdbridge.Status{
			Root:                       root,
			SocketPath:                 "/tmp/mwsd.sock",
			GraphSchemaVersion:         mwsdbridge.GraphSchemaVersion,
			DomainSchemaVersion:        mwsdbridge.DomainSchemaVersion,
			HarnessSchemaVersion:       mwsdbridge.HarnessSchemaVersion,
			OrchestrationSchemaVersion: mwsdbridge.OrchestrationSchemaVersion,
			DocumentCount:              1,
			RepositoryCount:            1,
		},
		Graph: mwsdbridge.GraphExport{
			SchemaVersion: mwsdbridge.GraphSchemaVersion,
			Root:          root,
			Documents: []mwsdbridge.Document{{
				Path:   "docs/CLI.md",
				ID:     "mws.cli",
				Title:  "CLI migration",
				Status: "in-progress",
				Owner:  "kim",
			}},
			Stats: mwsdbridge.GraphStats{
				DocumentCount: 1,
				NodeCount:     1,
				EdgeCount:     0,
			},
		},
		Domain: mwsdbridge.DomainExport{
			SchemaVersion: mwsdbridge.DomainSchemaVersion,
			Path:          "docs/domain.mws",
			Domain:        "macmini-workspace",
		},
		Harness: mwsdbridge.HarnessIndex{
			SchemaVersion:    mwsdbridge.HarnessSchemaVersion,
			RunCount:         1,
			TopDownCount:     1,
			BottomUpCount:    0,
			LastDirection:    "top-down",
			NextDirection:    "bottom-up",
			RecentDirections: []string{"top-down"},
		},
		Orchestration: mwsdbridge.OrchestrationSnapshot{
			SchemaVersion:          mwsdbridge.OrchestrationSchemaVersion,
			Root:                   root,
			DomainSchemaVersion:    mwsdbridge.DomainSchemaVersion,
			HarnessSchemaVersion:   mwsdbridge.HarnessSchemaVersion,
			Mode:                   "human-gated-provider-selection",
			DecisionGate:           "human-approval-required",
			DecisionBy:             []string{"codex"},
			DecisionLLMs:           []string{"codex"},
			ProviderCandidates:     []mwsdbridge.ProviderCandidate{{ID: "codex", SourceWorkflow: "provider-selection", Available: true, ApprovalRequired: true}},
			RecommendedProvider:    "codex",
			RecommendedDecisionLLM: "codex",
			NextAction: mwsdbridge.OrchestrationNextAction{
				Direction:             "bottom-up",
				CommandSurface:        "riido task queue",
				Reason:                "continue migration",
				RequiresHumanApproval: true,
			},
			TopDownCount:  1,
			BottomUpCount: 0,
			LastDirection: "top-down",
			Balanced:      true,
		},
		Projects: mwsdbridge.ProjectRegistry{
			SchemaVersion:   mwsdbridge.ProjectsSchemaVersion,
			Root:            root,
			RepositoryCount: 1,
			Repositories: []mwsdbridge.ProjectRepository{{
				Name:          "riido-daemon",
				Owner:         "teamswyg",
				Visibility:    "private",
				SSOTScope:     "docs",
				LocalPath:     root,
				Remote:        "https://github.com/teamswyg/riido-daemon",
				Role:          "daemon",
				LocalPresent:  true,
				GitPresent:    true,
				RemoteMatches: true,
			}},
		},
	}
}
