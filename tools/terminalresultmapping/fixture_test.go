package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFixtureRepo(t *testing.T, repo string) {
	t.Helper()
	mkdir(t, repo, "docs/20-domain/provider-runtime/adapter-draft-fields")
	mkdir(t, repo, "internal/agentbridge/supervisor")
	mkdir(t, repo, "internal/agentbridge")
	writeFile(t, repo, "docs/20-domain/provider-runtime/adapter-draft-fields/run-lifecycle.riido.json", fixtureManifestSource)
	writeFile(t, repo, "internal/agentbridge/result.go", fixtureResultSource)
	writeFile(t, repo, "internal/agentbridge/supervisor/terminal_result_draft.go", fixtureTerminalSource)
}

func mkdir(t *testing.T, repo, rel string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(repo, rel), 0o755); err != nil {
		t.Fatal(err)
	}
}
