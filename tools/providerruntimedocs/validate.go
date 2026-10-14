package main

const (
	manifestSchema = "riido-provider-runtime-docs.v1"
	pageSchema     = "riido-provider-runtime-page.v1"
)

func validateManifest(repo string, m manifest) ([]string, []sourceCheckResult) {
	var problems []string
	if m.SchemaVersion != manifestSchema {
		problems = append(problems, "unexpected schema_version")
	}
	if m.ID == "" || m.GeneratedDoc == "" || m.Workflow == "" || m.EvidenceArtifact == "" {
		problems = append(problems, "id, generated_doc, workflow, and evidence_artifact are required")
	}
	if len(m.Parts) < 8 || len(m.CompatibilityMarkers) == 0 || len(m.Assertions) == 0 {
		problems = append(problems, "parts, compatibility markers, and assertions are required")
	}
	problems = append(problems, validatePages(m.Pages)...)
	results, sourceProblems := validateSourceChecks(repo, m.SourceChecks)
	problems = append(problems, sourceProblems...)
	problems = append(problems, mustExist(repo, m.Workflow)...)
	return problems, results
}

func validatePages(pages []page) []string {
	if len(pages) != 3 {
		return []string{"three provider runtime wrapper pages are required"}
	}
	var problems []string
	seen := map[string]bool{}
	for _, page := range pages {
		problems = append(problems, validatePage(page, seen)...)
	}
	return problems
}
