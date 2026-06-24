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
		body, err := json.Marshal(candidate)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), "evidence_graph") {
			t.Fatalf("candidate JSON missing evidence_graph: %s", string(body))
		}
	}
}

func evidenceGraphIncomplete(graph evidenceGapCandidateGraph) bool {
	return graph.Observation == "" || graph.Hypothesis == "" ||
		graph.Change == "" || graph.Verifier == "" ||
		graph.Evidence == "" || graph.Decision == "" ||
		graph.NextLoop == ""
}
