package main

import "strings"

func buildEvidence(manifest Manifest, changed []string, results []bindingResult, problems []problem) Evidence {
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	return Evidence{
		SchemaVersion:  "riido-semantic-change-binding-evidence.v1",
		Status:         status,
		ManifestID:     manifest.ID,
		ChangedFiles:   changed,
		BusinessClaims: summarizeBusinessClaims(manifest, problems),
		Results:        results,
		ProblemCount:   len(problems),
		Problems:       problemMessages(problems),
	}
}

func summarizeBusinessClaims(manifest Manifest, problems []problem) businessClaimSummary {
	failed := businessProblemSet(manifest, problems)
	var summary businessClaimSummary
	for _, binding := range manifest.Bindings {
		if binding.ClaimClass != businessClaimClass {
			continue
		}
		summary.Count++
		summary.IDs = append(summary.IDs, binding.ID)
		if !failed[binding.ID] {
			summary.VerifiedCount++
		}
	}
	return summary
}

func businessProblemSet(manifest Manifest, problems []problem) map[string]bool {
	out := map[string]bool{}
	for _, binding := range manifest.Bindings {
		if binding.ClaimClass != businessClaimClass {
			continue
		}
		for _, problem := range problems {
			if strings.Contains(problem.Message, binding.ID) {
				out[binding.ID] = true
			}
		}
	}
	return out
}

func join(values []string) string {
	if len(values) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString(values[0])
	for _, value := range values[1:] {
		out.WriteString(", ")
		out.WriteString(value)
	}
	return out.String()
}
