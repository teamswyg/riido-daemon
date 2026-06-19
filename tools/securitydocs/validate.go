package main

const (
	manifestSchema = "riido-security-docs.v1"
	pageSchema     = "riido-security-page.v1"
)

func validateManifest(repo string, m manifest) ([]string, []sourceCheckResult) {
	var problems []string
	if m.SchemaVersion != manifestSchema {
		problems = append(problems, "unexpected schema_version")
	}
	if m.ID == "" || m.GeneratedDoc == "" || m.Workflow == "" || m.EvidenceArtifact == "" {
		problems = append(problems, "id, generated_doc, workflow, and evidence_artifact are required")
	}
	if len(m.Parts) < 3 || len(m.CompatibilityMarkers) == 0 || len(m.Assertions) == 0 {
		problems = append(problems, "parts, compatibility markers, and assertions are required")
	}
	if m.FullAccessHarnessTitle == "" || len(m.FullAccessHarness) == 0 || m.FullAccessHarnessCode == "" {
		problems = append(problems, "full access harness title, paragraphs, and code are required")
	}
	problems = append(problems, validatePages(m.Pages)...)
	results, sourceProblems := validateSourceChecks(repo, m.SourceChecks)
	problems = append(problems, sourceProblems...)
	problems = append(problems, mustExist(repo, m.Workflow)...)
	return problems, results
}

func validatePages(pages []page) []string {
	if len(pages) != 3 {
		return []string{"three security wrapper pages are required"}
	}
	var problems []string
	seen := map[string]bool{}
	for _, page := range pages {
		problems = append(problems, validatePage(page, seen)...)
	}
	return problems
}
