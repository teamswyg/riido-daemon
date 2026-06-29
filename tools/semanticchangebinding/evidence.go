package main

import "strings"

func buildEvidence(manifest Manifest, changed []string, results []bindingResult, problems []problem) Evidence {
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	return Evidence{
		SchemaVersion: "riido-semantic-change-binding-evidence.v1",
		Status:        status,
		ManifestID:    manifest.ID,
		ChangedFiles:  changed,
		Results:       results,
		ProblemCount:  len(problems),
		Problems:      problemMessages(problems),
	}
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
