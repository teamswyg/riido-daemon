package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDocEvidenceAndAbsentEdges(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	manifest := mustSessionManifest(t, repo)
	rendered := render(manifest)
	if got := checkDoc(options{Repo: repo}, manifest, rendered); len(got) != 0 {
		t.Fatalf("check disabled returned %#v", got)
	}
	assertSessionProblem(t, checkDoc(options{Repo: repo, CheckDoc: true}, manifest, rendered), "docs/session-lifecycle.md")
	if err := maybeWriteDoc(options{Repo: repo, WriteDoc: true}, manifest, rendered); err != nil {
		t.Fatal(err)
	}
	if got := checkDoc(options{Repo: repo, CheckDoc: true}, manifest, rendered); len(got) != 0 {
		t.Fatalf("fresh doc returned %#v", got)
	}
	writeFile(t, repo, manifest.GeneratedDoc, "stale")
	assertSessionProblem(t, checkDoc(options{Repo: repo, CheckDoc: true}, manifest, rendered), "generated doc drift")
	manifest.GeneratedDoc = "../bad.md"
	assertSessionProblem(t, checkDoc(options{Repo: repo, CheckDoc: true}, manifest, rendered), "unsafe repo path")
	if err := maybeWriteDoc(options{Repo: repo, WriteDoc: true}, manifest, rendered); err == nil {
		t.Fatal("expected unsafe write-doc failure")
	}
	writeFile(t, repo, "internal/agentbridge/nested/resume.go", "package nested\nfunc ResumeSession() {}\n")
	problems, absent := validateAbsent(repo, manifest.AbsentSurfaces)
	assertSessionProblem(t, problems, "unexpected surface")
	if len(absent) != 1 || absent[0].Pass || !strings.Contains(absent[0].Hits[0], "resume.go") {
		t.Fatalf("absent=%#v", absent)
	}
}

func TestIOAndValidationEdges(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	if _, err := loadManifest(repo, "../bad.json"); err == nil {
		t.Fatal("expected unsafe manifest path")
	}
	if _, err := readSource(repo, "../bad.go"); err == nil {
		t.Fatal("expected unsafe source path")
	}
	if err := writeJSON(filepath.Join(repo, "bad.json"), map[string]any{"bad": make(chan int)}); err == nil {
		t.Fatal("expected marshal failure")
	}
	manifest := mustSessionManifest(t, repo)
	manifest.Steps[0].SourceChecks = []string{"missing"}
	assertSessionProblem(t, validateStepReferences(manifest), "unknown source check")
	ev := buildEvidence(manifest, []problem{{"p"}}, nil, nil)
	if ev.Status != "failed" || len(ev.ProblemSummaries) != 1 {
		t.Fatalf("evidence=%#v", ev)
	}
}
