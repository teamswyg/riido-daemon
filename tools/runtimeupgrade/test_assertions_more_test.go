package main

import (
	"strings"
	"testing"
)

func mustRuntimeManifest(t *testing.T, repo string) Manifest {
	t.Helper()
	manifest, err := loadManifest(repo, defaultManifest)
	if err != nil {
		t.Fatal(err)
	}
	return manifest
}

func assertRuntimeProblem(t *testing.T, problems []problem, needle string) {
	t.Helper()
	for _, problem := range problems {
		if strings.Contains(problem.Message, needle) {
			return
		}
	}
	t.Fatalf("missing %q in %#v", needle, problems)
}
