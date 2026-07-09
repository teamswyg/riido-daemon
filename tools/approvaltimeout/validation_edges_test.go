package main

import (
	"path/filepath"
	"testing"
)

func TestValidateReportsManifestDrift(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	writeFile(t, repo, "docs/20-domain/provider-runtime/adapter-draft-fields/idle-watchdog.riido.json",
		`{"semantic_activity":["other"]}`)
	writeFile(t, repo, "docs/20-domain/provider-runtime/runtime-responsibility/provider-event-draft.riido.json",
		`{"mapped_events":[{"event_kind":"tool_approval_needed","event_type":"Other"}],"skipped_events":[]}`)
	problems, checks, sources := validate(repo, mustApprovalManifest(t, repo))
	if len(sources) != 5 || len(checks) != 3 {
		t.Fatalf("sources=%d checks=%d", len(sources), len(checks))
	}
	for _, check := range checks {
		if check.Pass {
			t.Fatalf("expected drift for %#v", check)
		}
	}
	for _, want := range []string{
		"approval_event_semantic_activity",
		"approval_event_maps_to_provider_draft",
		"timeout_not_provider_draft_owned",
	} {
		assertApprovalProblem(t, problems, want)
	}
}

func TestDocEvidenceAndIOEdges(t *testing.T) {
	dir := t.TempDir()
	manifest := Manifest{GeneratedDoc: "../bad.md", EvidenceArtifact: "artifact", Assertions: []string{"a"}}
	if got := checkDoc(options{Repo: dir}, manifest, ""); len(got) != 0 {
		t.Fatalf("check disabled returned %#v", got)
	}
	assertApprovalProblem(t, checkDoc(options{Repo: dir, CheckDoc: true}, manifest, ""), "unsafe repo path")
	if err := maybeWriteDoc(options{Repo: dir, WriteDoc: true}, manifest, "x"); err == nil {
		t.Fatal("expected unsafe write-doc failure")
	}
	manifest.GeneratedDoc = "docs/out.md"
	assertApprovalProblem(t, checkDoc(options{Repo: dir, CheckDoc: true}, manifest, ""), "docs/out.md")
	writeFile(t, dir, "docs/.keep", "")
	if err := maybeWriteDoc(options{Repo: dir, WriteDoc: true}, manifest, "fresh"); err != nil {
		t.Fatal(err)
	}
	if got := checkDoc(options{Repo: dir, CheckDoc: true}, manifest, "fresh"); len(got) != 0 {
		t.Fatalf("fresh doc returned %#v", got)
	}
	assertApprovalProblem(t, checkDoc(options{Repo: dir, CheckDoc: true}, manifest, "changed"), "generated doc drift")
	ev := buildEvidence(manifest, []problem{{"p"}}, nil, nil)
	if ev.Status != "failed" || len(ev.ProblemSummaries) != 1 {
		t.Fatalf("evidence=%#v", ev)
	}
	if err := writeJSON(filepath.Join(dir, "bad.json"), map[string]any{"bad": make(chan int)}); err == nil {
		t.Fatal("expected marshal failure")
	}
}
