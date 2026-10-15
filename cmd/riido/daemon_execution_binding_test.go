package main

import (
	"encoding/json"
	"net"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/execution"
)

func TestExecutionBindingIPCRequiresScopeAndPreservesOtherTask(t *testing.T) {
	store, err := execution.OpenBindingStore(filepath.Join(t.TempDir(), "bindings.json"))
	if err != nil {
		t.Fatal(err)
	}
	settings := daemonSettings{ExecutionBindingStore: store}
	call := func(req daemonRequest) map[string]string {
		t.Helper()
		client, server := net.Pipe()
		defer client.Close()
		go func() { defer server.Close(); writeExecutionBindingUpdate(server, req, settings) }()
		var result map[string]string
		if err := json.NewDecoder(client).Decode(&result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	binding := execution.Binding{WorkspaceID: "workspace-a", TaskID: "task-a", Repository: t.TempDir(), ChangedFiles: []string{"out"}, Checks: [][]string{{"git", "diff", "--check"}}}
	req := daemonRequest{Method: daemonMethodExecutionBindingUpdate, WorkspaceID: "workspace-a", TaskID: "task-b", ExecutionBinding: &binding}
	if call(req)["error"] == "" {
		t.Fatal("accepted mismatched task authority")
	}
	if len(store.Snapshot()) != 0 {
		t.Fatal("invalid request persisted")
	}
	req.TaskID = "task-a"
	if call(req)["status"] != "updated" {
		t.Fatal("valid binding not accepted")
	}
	snapshot := store.Snapshot()
	if call(daemonRequest{Method: daemonMethodExecutionBindingRemove, WorkspaceID: "workspace-a", TaskID: "task-b"})["status"] != "updated" {
		t.Fatal("idempotent remove failed")
	}
	if len(store.Snapshot()) != 1 || snapshot[0].TaskID != "task-a" {
		t.Fatal("other task changed")
	}
	if call(daemonRequest{Method: daemonMethodExecutionBindingRemove, WorkspaceID: "workspace-a", TaskID: "task-a"})["status"] != "updated" {
		t.Fatal("remove failed")
	}
	if len(store.Snapshot()) != 0 {
		t.Fatal("binding not removed")
	}
	if snapshot[0].Repository != binding.Repository {
		t.Fatal("prior execution snapshot changed")
	}
}
