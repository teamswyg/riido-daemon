package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadModelAndRunFromFixtureRepo(t *testing.T) {
	repo := t.TempDir()
	manifestPath := "docs/design.riido.json"
	writeDesignFixtureRepo(t, repo, manifestPath, testModel())
	evidenceOut := filepath.Join(repo, "out", "design-result.json")
	err := run(options{Repo: repo, Manifest: manifestPath, WriteDoc: true, CheckDoc: true, EvidenceOut: evidenceOut})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	body, err := os.ReadFile(evidenceOut)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	if !strings.Contains(string(body), `"status": "verified"`) {
		t.Fatalf("evidence body = %s", body)
	}
}

func TestRunReportsInvalidLoadedModel(t *testing.T) {
	repo := t.TempDir()
	m := testModel()
	m.Items = nil
	writeDesignFixtureRepo(t, repo, "docs/design.riido.json", m)
	err := run(options{Repo: repo, Manifest: "docs/design.riido.json"})
	if err == nil || !strings.Contains(err.Error(), "evidence items") {
		t.Fatalf("run err = %v", err)
	}
}

func hasAgentDesignProblem(problems []string, want string) bool {
	for _, problem := range problems {
		if strings.Contains(problem, want) {
			return true
		}
	}
	return false
}
