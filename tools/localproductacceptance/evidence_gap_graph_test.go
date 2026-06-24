package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEvidenceGapCandidatesCarryEvidenceGraph(t *testing.T) {
	got := evidenceGapCandidates([]scenario{{ID: "contract.api.bootstrap", Status: statusFailed}}, false, true, false)
	for _, candidate := range got {
		if evidenceGraphIncomplete(candidate.Graph) {
			t.Fatalf("candidate evidence graph incomplete: %+v", candidate)
		}
		if len(candidate.RequiredNextArtifacts) == 0 {
			t.Fatalf("candidate adoption artifacts missing: %+v", candidate)
		}
		body, err := json.Marshal(candidate)
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{"evidence_graph", "required_next_artifacts", "claim_binding"} {
			if !strings.Contains(string(body), want) {
				t.Fatalf("candidate JSON missing %q: %s", want, string(body))
			}
		}
	}
}

func evidenceGraphIncomplete(graph evidenceGapCandidateGraph) bool {
	return graph.Observation == "" || graph.Hypothesis == "" ||
		graph.Change == "" || graph.Verifier == "" ||
		graph.Evidence == "" || graph.Decision == "" ||
		graph.NextLoop == ""
}
