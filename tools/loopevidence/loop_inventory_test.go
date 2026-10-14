package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoopEvidenceFindsUnregisteredLoopFile(t *testing.T) {
	dir := t.TempDir()
	registered := filepath.Join(dir, "loops", "registered.riido.json")
	unregistered := filepath.Join(dir, "loops", "unregistered.riido.json")
	writeLoopFixture(t, registered)
	writeLoopFixture(t, unregistered)
	m := manifest{
		SchemaVersion:  "riido-loop-evidence.v1",
		ID:             "x",
		Title:          "X",
		GeneratedDoc:   "x.md",
		RequiredPhases: requiredPhases,
		LoopFiles:      []string{"loops/registered.riido.json"},
	}
	problems := validate(dir, m)
	if !containsLoopProblem(problems, "unregistered loop file loops/unregistered.riido.json") {
		t.Fatalf("problems = %#v", problems)
	}
}

func TestLoopEvidenceAllowsCompleteLoopInventory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "loops", "registered.riido.json")
	writeLoopFixture(t, path)
	loaded, err := loadLoopFile(path)
	if err != nil {
		t.Fatal(err)
	}
	m := manifest{
		SchemaVersion:  "riido-loop-evidence.v1",
		ID:             "x",
		Title:          "X",
		GeneratedDoc:   "x.md",
		RequiredPhases: requiredPhases,
		LoopFiles:      []string{"loops/registered.riido.json"},
		Loops:          []loop{loaded},
	}
	if problems := validate(dir, m); len(problems) != 0 {
		t.Fatalf("problems = %#v", problems)
	}
}

func containsLoopProblem(problems []string, want string) bool {
	for _, problem := range problems {
		if strings.Contains(problem, want) {
			return true
		}
	}
	return false
}
