package main

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratedChecksAndEvidenceEdges(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	manifest := mustBranchManifest(t, repo)
	rendered := renderedFiles{Doc: renderDoc(manifest), Script: renderScript(manifest)}
	checks := checkGenerated(options{Repo: repo, CheckDoc: true, CheckScript: true}, manifest, rendered)
	if len(checks) != 4 || checks[0].Pass || checks[1].Pass || !checks[2].Pass || !checks[3].Pass {
		t.Fatalf("checks=%#v", checks)
	}
	problems := scriptCheckProblems(checks)
	assertBranchProblem(t, problems, "docs/branch.md")
	evidence := buildEvidence(manifest, problems, checks, nil)
	if evidence.Status != "failed" || len(evidence.ProblemSummaries) != len(problems) {
		t.Fatalf("evidence=%#v problems=%#v", evidence, problems)
	}
	if got := checkContains(repo, "empty", "", "x"); !got.Pass || got.File != "" {
		t.Fatalf("empty contains check=%#v", got)
	}
}

func TestIOWriteAndExampleEdges(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	manifest := mustBranchManifest(t, repo)
	manifest.GeneratedDoc = "../bad.md"
	assertBranchProblem(t, maybeWrite(options{Repo: repo, WriteDoc: true}, manifest, renderedFiles{}), "unsafe repo path")
	manifest.GeneratedScript = "../bad.sh"
	problems, examples := runExamples(repo, manifest)
	if len(examples) != 0 {
		t.Fatalf("expected no examples on unsafe script, got %#v", examples)
	}
	assertBranchProblem(t, problems, "unsafe repo path")
	if _, err := loadManifest(repo, "../bad.json"); err == nil {
		t.Fatal("expected unsafe manifest path")
	}
	if _, err := readFile(repo, "../bad"); err == nil {
		t.Fatal("expected unsafe read path")
	}
	if err := writeJSON(filepath.Join(repo, "bad.json"), map[string]any{"bad": make(chan int)}); err == nil {
		t.Fatal("expected json marshal error")
	}
	if code := exitCode(errors.New("plain")); code != -1 {
		t.Fatalf("exitCode=%d", code)
	}
}

func mustBranchManifest(t *testing.T, repo string) Manifest {
	t.Helper()
	manifest, err := loadManifest(repo, defaultManifest)
	if err != nil {
		t.Fatal(err)
	}
	return manifest
}

func assertBranchProblem(t *testing.T, problems []problem, needle string) {
	t.Helper()
	for _, problem := range problems {
		if strings.Contains(problem.Message, needle) {
			return
		}
	}
	t.Fatalf("missing %q in %#v", needle, problems)
}
