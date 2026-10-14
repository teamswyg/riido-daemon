package main

import "fmt"

const schemaVersion = "riido-event-authority.v1"

func validate(repo string, manifest Manifest) ([]problem, []FieldCheck, []FieldCheck) {
	var problems []problem
	if manifest.SchemaVersion != schemaVersion {
		problems = append(problems, problem{fmt.Sprintf("schema_version must be %s", schemaVersion)})
	}
	problems = append(problems, validateRequired(manifest)...)
	fieldChecks, fieldProblems := validateDraftFields(repo, manifest)
	builderChecks, builderProblems := validateBuilderFields(repo, manifest)
	problems = append(problems, fieldProblems...)
	problems = append(problems, builderProblems...)
	return problems, fieldChecks, builderChecks
}

func validateRequired(manifest Manifest) []problem {
	values := []string{manifest.ID, manifest.Title, manifest.GeneratedDoc, manifest.Workflow, manifest.EvidenceArtifact, manifest.DraftSource, manifest.BuilderSource}
	var problems []problem
	for _, value := range values {
		if value == "" {
			problems = append(problems, problem{"manifest has empty required string"})
		}
	}
	return problems
}
