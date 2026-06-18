package workdir

import "testing"

func preparedTestWorkspace(t *testing.T, run string) (*FSAdapter, Workspace) {
	t.Helper()
	a := NewFSAdapter(t.TempDir())
	ws, err := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-" + run, Run: run})
	if err != nil {
		t.Fatal(err)
	}
	return a, ws
}

func injectedWorkspace(t *testing.T, cfg RuntimeConfig) Workspace {
	t.Helper()
	a, ws := preparedTestWorkspace(t, cfg.Provider)
	if err := a.InjectRuntimeConfig(ws, cfg); err != nil {
		t.Fatal(err)
	}
	return ws
}
