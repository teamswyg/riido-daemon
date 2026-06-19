package main

func buildEvidence(
	manifest Manifest,
	problems []problem,
	sources []SourceResult,
	reserved []ReservedRule,
) Evidence {
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	implemented, reservedCount := countRuleStatuses(manifest)
	return Evidence{
		SchemaVersion:    "riido-runtime-upgrade-flow-result.v1",
		ID:               manifest.ID,
		Status:           status,
		ImplementedRules: implemented,
		ReservedRules:    reservedCount,
		SourceChecks:     sources,
		Reserved:         reserved,
		Assertions:       manifest.Assertions,
		ProblemSummaries: problemMessages(problems),
		EvidenceArtifact: manifest.EvidenceArtifact,
	}
}

func countRuleStatuses(m Manifest) (int, int) {
	implemented, reserved := 0, 0
	for _, group := range ruleGroups(m) {
		for _, rule := range group.rules {
			switch rule.Status {
			case "reserved":
				reserved++
			case "implemented":
				implemented++
			}
		}
	}
	return implemented, reserved
}
