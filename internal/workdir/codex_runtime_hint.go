package workdir

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

// Hints locate candidates only. The stored session fingerprint remains authoritative.
func (a *FSAdapter) codexRuntimeHint(record NativeSession) NativeRuntimeHint {
	root, err := os.OpenRoot(a.root)
	if err != nil {
		return NativeRuntimeHint{}
	}
	defer root.Close()
	var hint NativeRuntimeHint
	if readRuntimeHeader(root, nativeSessionPath(record.Provider, record.SessionID)+".runtime", &hint) && hint.Version != "" {
		return hint
	}
	// Pre-hint bindings already have the detected CLI version in the first IR event.
	path, err := filepath.Rel(a.root, filepath.Join(filepath.Dir(record.Workdir), "ir", "events.jsonl"))
	if err != nil {
		return NativeRuntimeHint{}
	}
	var event struct{ ProviderVersion, CapabilityFingerprint string }
	if readRuntimeHeader(root, path, &event) && event.CapabilityFingerprint == record.Fingerprint {
		return NativeRuntimeHint{Version: event.ProviderVersion}
	}
	return NativeRuntimeHint{}
}

func readRuntimeHeader(root *os.Root, path string, value any) bool {
	info, err := root.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	file, err := root.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	return json.NewDecoder(io.LimitReader(file, 16384)).Decode(value) == nil
}
