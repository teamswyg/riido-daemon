package main

func buildEvidence(m manifest, docs []docClass, problems []string) evidence {
	counts := countDocs(docs)
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	if problems == nil {
		problems = []string{}
	}
	return evidence{
		SchemaVersion:    "riido-executable-knowledge-coverage-result.v1",
		ID:               m.ID,
		Status:           status,
		ScannedCount:     len(docs),
		GeneratedCount:   counts["generated"],
		DirectSSOTCount:  counts["direct_ssot"],
		ManualCount:      counts["manual_registered"],
		ManualGroups:     manualGroupIDs(m),
		ProblemSummaries: problems,
		EvidenceArtifact: m.EvidenceArtifact,
	}
}

func countDocs(docs []docClass) map[string]int {
	counts := map[string]int{}
	for _, doc := range docs {
		counts[doc.Kind]++
	}
	return counts
}
