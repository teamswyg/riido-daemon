package main

const (
	manifestSchema = "riido-readme-pages.v1"
	pageSchema     = "riido-readme-page.v1"
)

func validateManifest(repo string, m manifest) ([]string, []sourceCheckResult) {
	var problems []string
	if m.SchemaVersion != manifestSchema {
		problems = append(problems, "unexpected schema_version")
	}
	if m.ID == "" || m.Workflow == "" || m.EvidenceArtifact == "" {
		problems = append(problems, "id, workflow, and evidence_artifact are required")
	}
	if len(m.Assertions) == 0 {
		problems = append(problems, "assertions are required")
	}
	problems = append(problems, validatePages(m.Pages)...)
	results, sourceProblems := validateSourceChecks(repo, m.SourceChecks)
	problems = append(problems, sourceProblems...)
	problems = append(problems, mustExist(repo, m.Workflow)...)
	return problems, results
}

func validatePages(pages []page) []string {
	if len(pages) != 4 {
		return []string{"four readme handoff pages are required"}
	}
	var problems []string
	seen := map[string]bool{}
	for _, page := range pages {
		problems = append(problems, validatePage(page, seen)...)
	}
	return problems
}
