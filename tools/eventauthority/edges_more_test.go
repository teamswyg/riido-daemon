package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDocEvidenceAndIOEdges(t *testing.T) {
	repo := testRepo(t)
	manifest := validManifest("../bad.md")
	if got := checkDoc(options{Repo: repo}, manifest, ""); len(got) != 0 {
		t.Fatalf("check disabled returned %#v", got)
	}
	assertEventProblem(t, checkDoc(options{Repo: repo, CheckDoc: true}, manifest, ""), "path escapes repo")
	if err := maybeWriteDoc(options{Repo: repo, WriteDoc: true}, manifest, "x"); err == nil {
		t.Fatal("expected unsafe write-doc failure")
	}
	manifest.GeneratedDoc = "event-authority.md"
	assertEventProblem(t, checkDoc(options{Repo: repo, CheckDoc: true}, manifest, ""), "event-authority.md")
	if err := maybeWriteDoc(options{Repo: repo, WriteDoc: true}, manifest, "fresh"); err != nil {
		t.Fatal(err)
	}
	if got := checkDoc(options{Repo: repo, CheckDoc: true}, manifest, "fresh"); len(got) != 0 {
		t.Fatalf("fresh doc returned %#v", got)
	}
	assertEventProblem(t, checkDoc(options{Repo: repo, CheckDoc: true}, manifest, "changed"), "generated doc drift")
	ev := buildEvidence(manifest, []problem{{msg: "p"}}, nil, nil)
	if ev.Status != "failed" || len(ev.ProblemSummaries) != 1 {
		t.Fatalf("evidence=%#v", ev)
	}
	if problemError(nil) != nil || (problem{msg: "x"}).Error() != "x" {
		t.Fatal("unexpected problem helpers")
	}
}

func TestJSONPathAndParseEdges(t *testing.T) {
	repo := testRepo(t)
	if _, err := loadManifest(repo, "../bad.json"); err == nil {
		t.Fatal("expected escaped path failure")
	}
	mustWrite(t, filepath.Join(repo, "bad.json"), "{")
	if _, err := loadManifest(repo, "bad.json"); err == nil {
		t.Fatal("expected bad manifest json")
	}
	if err := writeJSON(filepath.Join(repo, "bad-out.json"), map[string]any{"bad": make(chan int)}); err == nil {
		t.Fatal("expected marshal failure")
	}
	manifest := validManifest("doc.md")
	manifest.DraftSource = "../draft.go"
	if _, err := draftFields(repo, manifest); err == nil {
		t.Fatal("expected unsafe draft path")
	}
	manifest = validManifest("doc.md")
	mustWrite(t, filepath.Join(repo, manifest.DraftSource), "package ingest\nfunc broken(")
	if _, err := draftFields(repo, manifest); err == nil {
		t.Fatal("expected draft parse error")
	}
	manifest = validManifest("doc.md")
	mustWrite(t, filepath.Join(repo, manifest.BuilderSource), "package ingest\nfunc broken(")
	if _, err := builderFields(repo, manifest); err == nil {
		t.Fatal("expected builder parse error")
	}
}

func assertEventProblem(t *testing.T, problems []problem, needle string) {
	t.Helper()
	for _, problem := range problems {
		if strings.Contains(problem.msg, needle) {
			return
		}
	}
	t.Fatalf("missing %q in %#v", needle, problems)
}
