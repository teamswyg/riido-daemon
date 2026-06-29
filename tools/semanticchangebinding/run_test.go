package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTriggerWithoutRequiredPeersFails(t *testing.T) {
	repo := fixtureRepo(t)
	err := run(context.Background(), options{
		Repo: repo, Manifest: manifestPath(),
		ChangedFiles: []string{"docs/30-architecture/closed-loop-maturity.dsl.json"},
	})
	if err == nil || !strings.Contains(err.Error(), "missing required semantic peers") {
		t.Fatalf("expected missing peer failure, got %v", err)
	}
}

func TestTriggerWithRequiredPeersPasses(t *testing.T) {
	repo := fixtureRepo(t)
	err := run(context.Background(), options{
		Repo: repo, Manifest: manifestPath(),
		ChangedFiles: []string{
			"docs/30-architecture/closed-loop-maturity.dsl.json",
			"tools/localproductacceptance/closed_loop_maturity_test.go",
			"tools/localproductacceptance/closed_loop_maturity.generated.json",
			".github/workflows/local-qa-runner.yml",
			"docs/30-architecture/loop-engineering/closed-loop-maturity.riido.json",
		},
	})
	if err != nil {
		t.Fatalf("expected clean binding, got %v", err)
	}
}

func TestGeneratedOnlyChangePasses(t *testing.T) {
	repo := fixtureRepo(t)
	err := run(context.Background(), options{
		Repo: repo, Manifest: manifestPath(),
		ChangedFiles: []string{"docs/30-architecture/loop-engineering.md"},
	})
	if err != nil {
		t.Fatalf("expected generated-only change to pass, got %v", err)
	}
}

func manifestPath() string {
	return "docs/30-architecture/semantic-change-bindings.riido.json"
}

func fixtureRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	copyFile(t, repo, manifestPath())
	for _, path := range fixturePaths() {
		writeFixtureFile(t, repo, path)
	}
	return repo
}

func copyFile(t *testing.T, repo, path string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("../..", path))
	if err != nil {
		t.Fatal(err)
	}
	writeFixtureFileWithData(t, repo, path, data)
}
