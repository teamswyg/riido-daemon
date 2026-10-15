package workdir

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

func nativeSessionFixture() NativeSession {
	return NativeSession{Schema: 1, Provider: "claude", SessionID: "native/../id", WorkspaceID: "workspace", TaskID: "task", AgentID: "agent", Fingerprint: "runtime-pin", GrantHash: "grant", OriginRunID: "first"}
}

func TestNativeSessionBindingSurvivesAdapterRestart(t *testing.T) {
	root := t.TempDir()
	adapter := NewFSAdapter(root)
	record := nativeSessionFixture()
	if err := adapter.PinNativeSession(record); err != nil {
		t.Fatal(err)
	}
	got, found, err := NewFSAdapter(root).LookupNativeSession(record.Provider, record.SessionID)
	if err != nil || !found {
		t.Fatalf("lookup=%+v %v %v", got, found, err)
	}
	if got.Workdir != filepath.Join(root, "workspace", "tasks", "task", "runs", "first", "workdir") {
		t.Fatal(got.Workdir)
	}
	got.Workdir = ""
	if got != record {
		t.Fatalf("binding=%+v", got)
	}
	info, err := os.Stat(filepath.Join(root, nativeSessionPath(record.Provider, record.SessionID)))
	if err != nil || (runtime.GOOS != "windows" && info.Mode().Perm() != 0o600) {
		t.Fatalf("permissions: %v %v", info, err)
	}
}

func TestNativeSessionBindingCannotBeReassigned(t *testing.T) {
	root := t.TempDir()
	original := nativeSessionFixture()
	if err := NewFSAdapter(root).PinNativeSession(original); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := range 6 {
		wg.Go(func() {
			record := original
			if i%2 == 0 {
				record.AgentID = "another-agent"
			}
			err := NewFSAdapter(root).PinNativeSession(record)
			if (err != nil) != (i%2 == 0) {
				t.Errorf("pin #%d: %v", i, err)
			}
		})
	}
	wg.Wait()
}

func TestNativeSessionBindingRejectsCorruptionAndTraversal(t *testing.T) {
	adapter := NewFSAdapter(t.TempDir())
	record := nativeSessionFixture()
	bad := record
	bad.OriginRunID = "../other"
	if adapter.PinNativeSession(bad) == nil {
		t.Fatal("accepted traversal")
	}
	if err := adapter.PinNativeSession(record); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(adapter.root, nativeSessionPath(record.Provider, record.SessionID))
	if err := os.WriteFile(path, []byte(`{"schema":1,"prompt":"not metadata"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := adapter.LookupNativeSession(record.Provider, record.SessionID); err == nil {
		t.Fatal("accepted corrupt binding")
	}
	if _, found, err := adapter.LookupNativeSession("codex", record.SessionID); err != nil || found {
		t.Fatalf("provider isolation: %v %v", found, err)
	}
}
