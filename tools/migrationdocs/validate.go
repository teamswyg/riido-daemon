package main

import "fmt"

const (
	manifestSchema = "riido-migration-docs.v1"
	pageSchema     = "riido-migration-page.v1"
)

func validateManifest(repo string, m manifest) ([]string, []sourceCheckResult) {
	var problems []string
	if m.SchemaVersion != manifestSchema {
		problems = append(problems, "unexpected schema_version")
	}
	if m.ID == "" || m.Workflow == "" || m.EvidenceArtifact == "" {
		problems = append(problems, "id, workflow, and evidence_artifact are required")
	}
	if m.ExpectedPageCount < 1 {
		problems = append(problems, "expected_page_count must be positive")
	}
	if len(m.Assertions) == 0 {
		problems = append(problems, "assertions are required")
	}
	problems = append(problems, validatePages(m.Pages, m.ExpectedPageCount)...)
	results, sourceProblems := validateSourceChecks(repo, m.SourceChecks)
	problems = append(problems, sourceProblems...)
	problems = append(problems, mustExist(repo, m.Workflow)...)
	return problems, results
}

func validatePages(pages []page, expectedCount int) []string {
	if len(pages) != expectedCount {
		msg := fmt.Sprintf("expected %d migration pages, got %d", expectedCount, len(pages))
		return []string{msg}
	}
	var problems []string
	seen := map[string]bool{}
	for _, page := range pages {
		problems = append(problems, validatePage(page, seen)...)
	}
	return problems
}
