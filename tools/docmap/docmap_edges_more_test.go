package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRequireNonEmptyAcceptsCompleteAndRejectsEmptyCollections(t *testing.T) {
	complete := manifest{
		ReadOrder: []readEntry{{Doc: "docs/a.md"}},
		Decisions: []decision{{Topic: "A", Docs: []string{"docs/a.md"}}},
		Repos:     []repo{{Repo: "riido-daemon", Responsibility: "daemon"}},
		Rules:     []string{"read docs before editing"},
	}
	if problems := requireNonEmpty(complete); len(problems) != 0 {
		t.Fatalf("complete manifest rejected: %#v", problems)
	}
	for name, m := range map[string]manifest{
		"read order": {Decisions: complete.Decisions, Repos: complete.Repos, Rules: complete.Rules},
		"decisions":  {ReadOrder: complete.ReadOrder, Repos: complete.Repos, Rules: complete.Rules},
		"repos":      {ReadOrder: complete.ReadOrder, Decisions: complete.Decisions, Rules: complete.Rules},
		"rules":      {ReadOrder: complete.ReadOrder, Decisions: complete.Decisions, Repos: complete.Repos},
	} {
		if problems := requireNonEmpty(m); len(problems) != 1 {
			t.Fatalf("%s gap should be rejected: %#v", name, problems)
		}
	}
}

func TestMaybeWriteOrCheckWritesAndDetectsDrift(t *testing.T) {
	root := t.TempDir()
	m := manifest{
		ID: "doc-map", Title: "Doc Map", Intro: "intro",
		GeneratedDocs: generatedDocs{
			Readme:      "docs/README.md",
			DocumentMap: "docs/readme/document-map.md",
		},
		ReadOrder: []readEntry{{Doc: "docs/a.md", Description: "A"}},
		Decisions: []decision{{Topic: "A", Docs: []string{"docs/a.md"}}},
		Repos:     []repo{{Repo: "repo", Responsibility: "docs"}},
		Rules:     []string{"rule"},
	}
	if err := maybeWriteOrCheck(root, m, false, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "docs", "README.md")); !os.IsNotExist(err) {
		t.Fatalf("disabled write should not create docs, err=%v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "docs", "readme"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := maybeWriteOrCheck(root, m, true, true); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "README.md"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := maybeWriteOrCheck(root, m, false, true)
	if err == nil || !strings.Contains(err.Error(), "generated doc drift") {
		t.Fatalf("expected generated doc drift, got %v", err)
	}
}
