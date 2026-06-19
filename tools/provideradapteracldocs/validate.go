package main

const (
	manifestSchema = "riido-provider-adapter-acl-docs.v1"
	detailSchema   = "riido-provider-adapter-acl-detail.v1"
)

func validateManifest(repo string, m manifest) ([]string, []sourceCheckResult) {
	var problems []string
	if m.SchemaVersion != manifestSchema {
		problems = append(problems, "unexpected schema_version")
	}
	if m.ID == "" || m.GeneratedDoc == "" || m.Workflow == "" || m.EvidenceArtifact == "" {
		problems = append(problems, "id, generated_doc, workflow, and evidence_artifact are required")
	}
	if len(m.DetailPages) != 5 || len(m.Details) != 5 {
		problems = append(problems, "five generated detail pages are required")
	}
	if len(m.Assertions) == 0 || len(m.SourceChecks) == 0 || len(m.RelatedPages) == 0 {
		problems = append(problems, "assertions, source checks, and related pages are required")
	}
	problems = append(problems, validateDetails(m)...)
	results, sourceProblems := validateSourceChecks(repo, m.SourceChecks)
	problems = append(problems, sourceProblems...)
	problems = append(problems, mustExist(repo, m.Workflow)...)
	return problems, results
}

func validateDetails(m manifest) []string {
	var problems []string
	seen := map[string]bool{}
	for _, detail := range m.Details {
		if detail.SchemaVersion != detailSchema {
			problems = append(problems, "unexpected detail schema_version: "+detail.ID)
		}
		if detail.ID == "" || detail.Title == "" || detail.GeneratedDoc == "" || len(detail.Blocks) == 0 {
			problems = append(problems, "detail id, title, generated_doc, and blocks are required")
		}
		if seen[detail.ID] {
			problems = append(problems, "duplicate detail id "+detail.ID)
		}
		seen[detail.ID] = true
		problems = append(problems, validateBlocks(detail)...)
	}
	return problems
}
