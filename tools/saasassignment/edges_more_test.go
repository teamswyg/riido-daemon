package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDocEvidenceAndAbsentEdges(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	manifest := mustSaaSManifest(t, repo)
	docs := renderedDocs(manifest)
	if got := checkDocs(options{Repo: repo}, docs); len(got) != 0 {
		t.Fatalf("check disabled returned %#v", got)
	}
	assertSaaSProblem(t, checkDocs(options{Repo: repo, CheckDoc: true}, docs), "docs/domain.md")
	if err := maybeWriteDocs(options{Repo: repo, WriteDoc: true}, docs); err != nil {
		t.Fatal(err)
	}
	if got := checkDocs(options{Repo: repo, CheckDoc: true}, docs); len(got) != 0 {
		t.Fatalf("fresh docs returned %#v", got)
	}
	writeFile(t, repo, manifest.GeneratedDoc, "stale")
	assertSaaSProblem(t, checkDocs(options{Repo: repo, CheckDoc: true}, docs), "generated doc drift")
	badDocs := map[string]string{"../bad.md": "x"}
	assertSaaSProblem(t, checkDocs(options{Repo: repo, CheckDoc: true}, badDocs), "unsafe repo path")
	if err := maybeWriteDocs(options{Repo: repo, WriteDoc: true}, badDocs); err == nil {
		t.Fatal("expected unsafe write-doc failure")
	}
	writeFile(t, repo, "internal/nested/hit.go", "package nested\nconst _ = \"thread-progress\"\n")
	problems, absent := validateAbsent(repo, manifest.AbsentSurfaces)
	assertSaaSProblem(t, problems, "unexpected surface")
	if len(absent) != 1 || absent[0].Pass || !strings.Contains(absent[0].Hits[0], "internal/nested/hit.go") {
		t.Fatalf("absent=%#v", absent)
	}
}

func TestManifestIOAndEvidenceEdges(t *testing.T) {
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
	manifest := mustSaaSManifest(t, repo)
	manifest.SchemaVersion = ""
	manifest.Facts = []Fact{{Name: "implemented", Status: "implemented"}}
	manifest.SourceChecks = []SourceCheck{{Name: "dupe"}, {Name: "dupe"}}
	problems, _, _ := validate(repo, manifest)
	for _, want := range []string{"required field", "no source checks", "duplicate source check"} {
		assertSaaSProblem(t, problems, want)
	}
	ev := buildEvidence(manifest, []problem{{"p"}}, nil, nil)
	if ev.Status != "failed" || len(ev.ProblemSummaries) != 1 {
		t.Fatalf("evidence=%#v", ev)
	}
}
