package workdir

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCodexRuntimeHintPreservesLegacyIndex(t *testing.T) {
	root := t.TempDir()
	store := NewFSAdapter(root)
	record := nativeSessionFixture()
	record.Provider = "codex"
	record.RuntimeHint = NativeRuntimeHint{Executable: "/host/codex", Version: "original"}
	if err := store.PinNativeSession(record); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, nativeSessionPath("codex", record.SessionID)))
	if err != nil {
		t.Fatal(err)
	}
	var index map[string]any
	if err := json.Unmarshal(data, &index); err != nil {
		t.Fatal(err)
	}
	if _, exists := index["RuntimeHint"]; exists {
		t.Fatal("changed legacy index")
	}
	got, found, err := NewFSAdapter(root).LookupNativeSession("codex", record.SessionID)
	if err != nil || !found || got.RuntimeHint != record.RuntimeHint {
		t.Fatalf("hint=%+v %v", got.RuntimeHint, err)
	}
}

func TestCodexLegacyHintRequiresMatchingOriginFingerprint(t *testing.T) {
	root := t.TempDir()
	store := NewFSAdapter(root)
	record := nativeSessionFixture()
	record.Provider = "codex"
	if err := store.PinNativeSession(record); err != nil {
		t.Fatal(err)
	}
	got, _, _ := store.LookupNativeSession("codex", record.SessionID)
	path := filepath.Join(filepath.Dir(got.Workdir), "ir", "events.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, fingerprint := range []string{record.Fingerprint, "foreign"} {
		data, _ := json.Marshal(map[string]string{"CapabilityFingerprint": fingerprint, "ProviderVersion": "original"})
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
		got, _, err := store.LookupNativeSession("codex", record.SessionID)
		if err != nil {
			t.Fatal(err)
		}
		if (got.RuntimeHint.Version != "") != (fingerprint == record.Fingerprint) {
			t.Fatalf("hint=%+v", got.RuntimeHint)
		}
	}
}
