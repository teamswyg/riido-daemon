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
