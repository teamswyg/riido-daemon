package main

import (
	"strings"
	"testing"
)

func TestValidateHeaderAcceptsCompleteManifestAndReportsGaps(t *testing.T) {
	complete := manifest{
		SchemaVersion:  "riido-loop-evidence.v1",
		ID:             "loops",
		Title:          "Loops",
		GeneratedDoc:   "docs/loops.md",
		RequiredPhases: requiredPhases,
	}
	if problems := validateHeader(complete); len(problems) != 0 {
		t.Fatalf("complete header rejected: %#v", problems)
	}

	missing := complete
	missing.SchemaVersion = "old"
	missing.ID = ""
	missing.RequiredPhases = []string{"observe"}
	problems := validateHeader(missing)
	if len(problems) != 3 {
		t.Fatalf("expected schema, required field, and phase problems: %#v", problems)
	}
}

func TestValidateGapRequiresEscalationArtifact(t *testing.T) {
	valid := gap{
		ID:                   "gap",
		Owner:                "qa",
		CurrentHandling:      "manual review",
		RequiredNextArtifact: "tools/loopevidence",
	}
	if problems := validateGap(valid); len(problems) != 0 {
		t.Fatalf("valid gap rejected: %#v", problems)
	}
	for name, item := range map[string]gap{
		"missing id":       {Owner: "qa", CurrentHandling: "manual", RequiredNextArtifact: "x"},
		"missing owner":    {ID: "gap", CurrentHandling: "manual", RequiredNextArtifact: "x"},
		"missing handling": {ID: "gap", Owner: "qa", RequiredNextArtifact: "x"},
		"missing artifact": {ID: "gap", Owner: "qa", CurrentHandling: "manual"},
	} {
		problems := validateGap(item)
		if len(problems) != 1 || !strings.Contains(problems[0], "required_next_artifact") {
			t.Fatalf("%s was not rejected correctly: %#v", name, problems)
		}
	}
}
