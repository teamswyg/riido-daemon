package main

import "fmt"

func validateManifests(
	manifest Manifest,
	semantic SemanticActivityManifest,
	draft ProviderDraftManifest,
) ([]problem, []ManifestCheck) {
	checks := []ManifestCheck{
		semanticCheck(manifest.ApprovalEvent.EventKind, semantic.SemanticActivity),
		mappedDraftCheck(manifest.ApprovalEvent, draft.MappedEvents),
		skippedTimeoutCheck(manifest.TimeoutEvent.EventKind, draft.SkippedEvents),
	}
	var problems []problem
	for _, check := range checks {
		if !check.Pass {
			problems = append(problems, problem{fmt.Sprintf("manifest drift: %s", check.Name)})
		}
	}
	return problems, checks
}

func semanticCheck(kind string, semantic []string) ManifestCheck {
	check := ManifestCheck{Name: "approval_event_semantic_activity", Expected: kind}
	for _, candidate := range semantic {
		if candidate == kind {
			check.Actual, check.Pass = candidate, true
			return check
		}
	}
	return check
}
