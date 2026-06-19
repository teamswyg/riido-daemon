package main

import (
	"path/filepath"
	"testing"
)

func TestWorkspaceDocsAreGenerated(t *testing.T) {
	repo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	m, err := readManifest(filepath.Join(repo, defaultManifest))
	if err != nil {
		t.Fatal(err)
	}
	problems, _ := validateManifest(repo, m)
	problems = append(problems, checkDocs(repo, m)...)
	if len(problems) > 0 {
		t.Fatalf("workspace docs problems: %v", problems)
	}
}
