package childreg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegistryPersistsLiveSet(t *testing.T) {
	path := filepath.Join(t.TempDir(), "daemon-children.pids")
	r := New(path)
	r.OnSpawn(111)
	r.OnSpawn(222)
	r.OnExit(111)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}
	got := strings.Fields(string(data))
	if len(got) != 1 || got[0] != "222" {
		t.Fatalf("registry file = %q, want only 222", string(data))
	}
}

func TestRegistryIgnoresInvalidPid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "daemon-children.pids")
	r := New(path)
	r.OnSpawn(0)
	r.OnSpawn(-5)
	if data, _ := os.ReadFile(path); len(strings.Fields(string(data))) != 0 {
		t.Fatalf("non-positive pids should not be recorded, got %q", string(data))
	}
}

func TestReapOrphansMissingFileIsClean(t *testing.T) {
	reaped, err := ReapOrphans(filepath.Join(t.TempDir(), "absent.pids"))
	if err != nil || reaped != 0 {
		t.Fatalf("missing file: reaped=%d err=%v, want 0/nil", reaped, err)
	}
}

func TestReapOrphansEmptyPath(t *testing.T) {
	if reaped, err := ReapOrphans("  "); err != nil || reaped != 0 {
		t.Fatalf("empty path: reaped=%d err=%v, want 0/nil", reaped, err)
	}
}
