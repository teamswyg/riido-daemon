package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
)

func TestGitChangedFilesPrefersStagedChanges(t *testing.T) {
	repo := initGitFixture(t)
	writeGitFile(t, repo, "tracked.txt", "changed")
	writeGitFile(t, repo, "staged.txt", "new")
	gitFixture(t, repo, "add", "staged.txt")
	writeGitFile(t, repo, "untracked.txt", "new")
	got := gitChangedFiles(repo)
	if len(got) != 1 || got[0] != "staged.txt" {
		t.Fatalf("changed files=%v", got)
	}
}

func TestGitChangedFilesIncludesUnstagedAndUntracked(t *testing.T) {
	repo := initGitFixture(t)
	writeGitFile(t, repo, "tracked.txt", "changed")
	writeGitFile(t, repo, "untracked.txt", "new")
	got := gitChangedFiles(repo)
	if !slices.Contains(got, "tracked.txt") || !slices.Contains(got, "untracked.txt") {
		t.Fatalf("changed files=%v", got)
	}
}

func TestSplitLinesAndDedupeNormalizeGitOutput(t *testing.T) {
	lines := splitLines(" a.go \n\nb.go\n")
	if len(lines) != 2 || lines[0] != "a.go" || lines[1] != "b.go" {
		t.Fatalf("lines=%v", lines)
	}
	got := dedupe([]string{"a.go", "a.go", "b.go"})
	if len(got) != 2 || got[0] != "a.go" || got[1] != "b.go" {
		t.Fatalf("dedupe=%v", got)
	}
}

func initGitFixture(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	gitFixture(t, repo, "init")
	writeGitFile(t, repo, "tracked.txt", "base")
	gitFixture(t, repo, "add", "tracked.txt")
	gitFixture(t, repo, "-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "-m", "init")
	return repo
}

func writeGitFile(t *testing.T, repo, path, body string) {
	t.Helper()
	full := filepath.Join(repo, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func gitFixture(t *testing.T, repo string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}
