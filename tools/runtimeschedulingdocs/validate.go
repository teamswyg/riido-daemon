package main

const (
	manifestSchema = "riido-runtime-scheduling-docs.v1"
	coreSchema     = "riido-runtime-scheduling-core.v1"
)

func validateManifest(repo string, m manifest) ([]string, []sourceCheckResult) {
	var problems []string
	if m.SchemaVersion != manifestSchema || m.Core.SchemaVersion != coreSchema {
		problems = append(problems, "unexpected schema_version")
	}
	if m.ID == "" || m.GeneratedDoc == "" || m.Workflow == "" || m.EvidenceArtifact == "" {
		problems = append(problems, "id, generated_doc, workflow, and evidence_artifact are required")
	}
	problems = append(problems, validateLinks("parts", m.Parts, 2)...)
	problems = append(problems, validateIndex(m.InvariantsIndex)...)
	problems = append(problems, validateCore(m.Core)...)
	problems = append(problems, validateInvariantChecks(m)...)
	results, sourceProblems := validateSourceChecks(repo, m.SourceChecks)
	problems = append(problems, sourceProblems...)
	problems = append(problems, mustExist(repo, m.Workflow)...)
	return problems, results
}

func validateIndex(index indexDoc) []string {
	var problems []string
	if index.Title == "" || index.GeneratedDoc == "" || len(index.Summary) == 0 {
		problems = append(problems, "invariants index title, generated_doc, and summary are required")
	}
	return append(problems, validateLinks("invariants parts", index.Parts, 4)...)
}
