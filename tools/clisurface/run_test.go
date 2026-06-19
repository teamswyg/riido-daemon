package main

import "testing"

func TestRunAgainstRepositoryEvidence(t *testing.T) {
	err := run(options{Repo: "../..", Manifest: defaultManifest})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRenderIncludesGeneratedHeader(t *testing.T) {
	manifest := Manifest{Title: "CLI", EvidenceArtifact: "artifact"}
	body := renderMarkdown(manifest)
	if body == "" || body[0] != '#' {
		t.Fatalf("unexpected rendered body: %q", body)
	}
}
