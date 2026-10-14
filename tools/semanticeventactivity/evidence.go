package main

import "sort"

func buildEvidence(manifest Manifest, problems []problem) Evidence {
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	return Evidence{
		SchemaVersion:    "riido-semantic-event-activity-result.v1",
		ID:               manifest.ID,
		Status:           status,
		Classifications:  classifications(manifest),
		Assertions:       append([]string(nil), manifest.Assertions...),
		ProblemSummaries: problemMessages(problems),
		EvidenceArtifact: manifest.EvidenceArtifact,
	}
}

func classifications(manifest Manifest) []Classification {
	kinds, _ := manifestKindMap(manifest)
	out := make([]Classification, 0, len(kinds))
	for kind, manifestSemantic := range kinds {
		runtimeSemantic, ok := runtimeKinds()[kind]
		out = append(out, Classification{Kind: kind, Manifest: category(manifestSemantic), Runtime: category(runtimeSemantic), Confirmed: ok && manifestSemantic == runtimeSemantic})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Kind < out[j].Kind })
	return out
}
