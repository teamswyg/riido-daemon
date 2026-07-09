package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestProblemHelpersExposeFirstProblem(t *testing.T) {
	if err := problemError(nil); err != nil {
		t.Fatalf("empty problem list should not fail: %v", err)
	}
	problems := []problem{{"first"}, {"second"}}
	if got := problemError(problems).Error(); got != "first" {
		t.Fatalf("expected first problem, got %q", got)
	}
	messages := problemMessages(problems)
	if len(messages) != 2 || messages[0] != "first" || messages[1] != "second" {
		t.Fatalf("unexpected messages: %#v", messages)
	}
}

func TestCleanRepoPathGuardsRepoBoundary(t *testing.T) {
	repo := t.TempDir()
	cases := map[string]bool{
		"doc.md":  true,
		".":       false,
		"..":      false,
		"../x.md": false,
	}
	for path, ok := range cases {
		got, err := cleanRepoPath(repo, path)
		if ok && (err != nil || got != filepath.Join(repo, path)) {
			t.Fatalf("expected clean path for %q, got %q err %v", path, got, err)
		}
		if !ok && err == nil {
			t.Fatalf("expected %q to be rejected", path)
		}
	}
	if _, err := cleanRepoPath(repo, filepath.Join(repo, "doc.md")); err == nil {
		t.Fatalf("expected absolute path to be rejected")
	}
}

func TestCheckDocReportsMissingDriftAndFreshDoc(t *testing.T) {
	repo := t.TempDir()
	manifest := validManifest("doc.md")
	rendered := render(manifest)
	if problems := checkDoc(options{Repo: repo}, manifest, rendered); len(problems) != 0 {
		t.Fatalf("disabled doc check should not report problems: %#v", problems)
	}
	problems := checkDoc(options{Repo: repo, CheckDoc: true}, manifest, rendered)
	if len(problems) != 1 || !strings.Contains(problems[0].msg, "no such file") {
		t.Fatalf("expected missing doc problem, got %#v", problems)
	}
	mustWrite(t, filepath.Join(repo, "doc.md"), "stale")
	problems = checkDoc(options{Repo: repo, CheckDoc: true}, manifest, rendered)
	if len(problems) != 1 || !strings.Contains(problems[0].msg, "drift") {
		t.Fatalf("expected drift problem, got %#v", problems)
	}
	mustWrite(t, filepath.Join(repo, "doc.md"), rendered)
	if problems := checkDoc(options{Repo: repo, CheckDoc: true}, manifest, rendered); len(problems) != 0 {
		t.Fatalf("fresh doc should pass: %#v", problems)
	}
}
