package main

import "strings"

const evidenceSchema = "riido-daemon-package-workflow-evidence-result.v1"

func buildEvidence(m manifest, spec workflowSpec, workflowText string) evidence {
	results := make([]fragmentResult, 0, len(spec.RequiredFragments))
	matched := 0
	for _, fragment := range spec.RequiredFragments {
		found := strings.Contains(workflowText, fragment)
		if found {
			matched++
		}
		results = append(results, fragmentResult{Value: fragment, Found: found})
	}
	status := "failed"
	if matched == len(spec.RequiredFragments) {
		status = "verified"
	}
	return evidence{
		SchemaVersion:    evidenceSchema,
		ID:               spec.ID,
		Status:           status,
		Workflow:         spec.Workflow,
		EvidenceArtifact: spec.EvidenceArtifact,
		MatchedCount:     matched,
		RequiredCount:    len(spec.RequiredFragments),
		Fragments:        results,
		LoopSource:       m.LoopSource,
	}
}
