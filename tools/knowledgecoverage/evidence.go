package main

func buildEvidence(root string, m manifest, docs []docClass, problems []string) evidence {
	counts := countDocs(docs)
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	if problems == nil {
		problems = []string{}
	}
	inventory, err := scanManifestInventory(root)
	if err != nil {
		status = "failed"
		problems = append(problems, "manifest inventory scan failed: "+err.Error())
	}
	return evidence{
		SchemaVersion:     "riido-executable-knowledge-coverage-result.v1",
		ID:                m.ID,
		Status:            status,
		ScannedCount:      len(docs),
		GeneratedCount:    counts["generated"],
		DirectSSOTCount:   counts["direct_ssot"],
		ManualCount:       counts["manual_registered"],
		ManualGroups:      manualGroupIDs(m),
		ManualByGroup:     manualCountsByGroup(docs),
		ManualTopDirs:     manualTopDirs(docs, 12),
		ManualSamples:     manualSamples(docs, 5),
		ManifestInventory: inventory,
		ProblemSummaries:  problems,
		EvidenceArtifact:  m.EvidenceArtifact,
		Loop:              m.Loop,
	}
}

func countDocs(docs []docClass) map[string]int {
	counts := map[string]int{}
	for _, doc := range docs {
		counts[doc.Kind]++
	}
	return counts
}

func manualCountsByGroup(docs []docClass) map[string]int {
	counts := map[string]int{}
	for _, doc := range docs {
		if doc.Kind == "manual_registered" {
			counts[doc.Group]++
		}
	}
	return counts
}
