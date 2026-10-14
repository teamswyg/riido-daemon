package figmaboundary

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

const (
	boundaryManifestRelPath = "docs/30-architecture/figma-ai-agent-daemon-boundary.riido.json"
	boundaryHumanDocRelPath = "docs/30-architecture/figma-ai-agent-daemon-boundary.md"
)

func loadBoundaryManifest(t *testing.T) boundaryManifest {
	t.Helper()
	var manifest boundaryManifest
	path := filepath.Join(repoRoot(t), boundaryManifestRelPath)
	if err := json.Unmarshal(mustReadFile(t, path), &manifest); err != nil {
		t.Fatalf("decode daemon Figma boundary manifest: %v", err)
	}
	manifest.Entries = append(manifest.Entries, loadBoundaryEntryFiles(t, path, manifest.EntryFiles)...)
	return manifest
}

func loadBoundaryEntryFiles(t *testing.T, manifestPath string, files []string) []boundaryEntry {
	t.Helper()
	entries := []boundaryEntry{}
	for _, file := range files {
		var loaded []boundaryEntry
		path := filepath.Join(filepath.Dir(manifestPath), filepath.FromSlash(file))
		if err := json.Unmarshal(mustReadFile(t, path), &loaded); err != nil {
			t.Fatalf("decode boundary entry file %s: %v", file, err)
		}
		entries = append(entries, loaded...)
	}
	return entries
}

func boundaryEntriesByNodeID(t *testing.T, manifest boundaryManifest) map[string]boundaryEntry {
	t.Helper()
	entries := map[string]boundaryEntry{}
	for _, entry := range manifest.Entries {
		requireValidBoundaryEntry(t, entry)
		entries[entry.NodeID] = entry
	}
	return entries
}

func requireValidBoundaryEntry(t *testing.T, entry boundaryEntry) {
	t.Helper()
	if entry.NodeID == "" || entry.Name == "" || entry.DaemonScope == "" {
		t.Fatalf("entry must include node, name, and daemon_scope: %#v", entry)
	}
	if len(entry.UpstreamOwner) == 0 {
		t.Fatalf("entry %s must name upstream owners", entry.NodeID)
	}
	if entry.DaemonConsumedFacts == nil {
		t.Fatalf("entry %s must include daemon_consumed_facts, even when empty", entry.NodeID)
	}
	if len(entry.ClientOwnedFacts) == 0 {
		t.Fatalf("entry %s must separate client_owned_facts", entry.NodeID)
	}
}
