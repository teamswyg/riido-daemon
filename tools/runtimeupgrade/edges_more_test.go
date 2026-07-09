package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocEvidenceAndIOEdges(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	manifest := mustRuntimeManifest(t, repo)
	if got := checkDoc(options{Repo: repo}, manifest, ""); len(got) != 0 {
		t.Fatalf("check disabled returned %#v", got)
	}
	manifest.GeneratedDoc = "../bad.md"
	assertRuntimeProblem(t, checkDoc(options{Repo: repo, CheckDoc: true}, manifest, ""), "unsafe repo path")
	if err := maybeWriteDoc(options{Repo: repo, WriteDoc: true}, manifest, "x"); err == nil {
		t.Fatal("expected unsafe write-doc failure")
	}
	manifest.GeneratedDoc = "docs/runtime-upgrade.md"
	assertRuntimeProblem(t, checkDoc(options{Repo: repo, CheckDoc: true}, manifest, ""), "docs/runtime-upgrade.md")
	if err := maybeWriteDoc(options{Repo: repo, WriteDoc: true}, manifest, "fresh"); err != nil {
		t.Fatal(err)
	}
	if got := checkDoc(options{Repo: repo, CheckDoc: true}, manifest, "fresh"); len(got) != 0 {
		t.Fatalf("fresh doc returned %#v", got)
	}
	assertRuntimeProblem(t, checkDoc(options{Repo: repo, CheckDoc: true}, manifest, "changed"), "generated doc drift")
	if err := writeJSON(filepath.Join(repo, "bad.json"), map[string]any{"bad": make(chan int)}); err == nil {
		t.Fatal("expected marshal failure")
	}
	ev := buildEvidence(manifest, []problem{{"p"}}, nil, nil, nil)
	if ev.Status != "failed" || len(ev.ProblemSummaries) != 1 {
		t.Fatalf("evidence=%#v", ev)
	}
}

func TestAbsentDirectoryScanAndRenderEdges(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, repo, "claims/nested.txt", "RuntimePinViolated")
	problems, checks := validateAbsent(repo, []AbsentSurface{{
		Name: "claim", Scope: []string{"claims"}, Tokens: []string{"RuntimePinViolated"},
	}})
	assertRuntimeProblem(t, problems, "unexpected claim")
	if len(checks) != 1 || checks[0].Pass || !strings.Contains(checks[0].Hits[0], "claims/nested.txt") {
		t.Fatalf("checks=%#v", checks)
	}
	problems, _ = validateAbsent(repo, []AbsentSurface{{Name: "bad", Scope: []string{"../bad"}, Tokens: []string{"x"}}})
	assertRuntimeProblem(t, problems, "unsafe repo path")
	var b bytes.Buffer
	writeAbsentSurfaces(&b, nil)
	if b.Len() != 0 {
		t.Fatalf("absent render=%q", b.String())
	}
	if ruleDecisions(Rule{}) != "" || !strings.Contains(ruleDecisions(Rule{DecisionRefs: []string{"A"}}), "`A`") {
		t.Fatal("unexpected rule decisions")
	}
}
