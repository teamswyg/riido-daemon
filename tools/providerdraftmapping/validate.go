package main

import "fmt"

const schemaVersion = "riido-provider-draft-mapping.v1"

func validate(repo string, manifest Manifest) ([]problem, []MappingCheck, []CoverageCheck) {
	var problems []problem
	if manifest.SchemaVersion != schemaVersion {
		problems = append(problems, problem{fmt.Sprintf("schema_version must be %s", schemaVersion)})
	}
	problems = append(problems, validateRequired(manifest)...)
	source, err := sourceMapping(repo, manifest)
	if err != nil {
		return append(problems, problem{err.Error()}), nil, nil
	}
	mappingChecks, mappingProblems := validateMappings(manifest, source)
	coverageChecks, coverageProblems := validateCoverage(manifest)
	problems = append(problems, mappingProblems...)
	problems = append(problems, coverageProblems...)
	return problems, mappingChecks, coverageChecks
}

func validateRequired(manifest Manifest) []problem {
	values := []string{manifest.ID, manifest.Title, manifest.GeneratedDoc, manifest.Workflow, manifest.EvidenceArtifact, manifest.Source}
	var problems []problem
	for _, value := range values {
		if value == "" {
			problems = append(problems, problem{"manifest has empty required string"})
		}
	}
	return problems
}
