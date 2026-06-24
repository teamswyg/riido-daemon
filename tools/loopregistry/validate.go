package main

import "slices"

func validateRegistry(root string, reg registry) []string {
	var problems []string
	problems = append(problems, validateHeader(root, reg)...)
	problems = append(problems, validateLoop(reg.Loop, "registry loop")...)
	if len(reg.Loops) == 0 {
		problems = append(problems, "loops are required")
	}
	for _, item := range reg.Loops {
		problems = append(problems, validateLoopEntry(item)...)
	}
	if len(reg.BusinessClaims) == 0 {
		problems = append(problems, "business_claims are required")
	}
	problems = append(problems, validateClaims(root, reg.BusinessClaims)...)
	problems = append(problems, validateCIBindings(root, reg)...)
	return problems
}

func validateHeader(root string, reg registry) []string {
	var problems []string
	if reg.SchemaVersion != schemaVersion {
		problems = append(problems, "schema_version must be "+schemaVersion)
	}
	required := []string{
		reg.ID, reg.Title, reg.GeneratedDoc, reg.Workflow,
		reg.EvidenceArtifact, reg.PrecommitHook, reg.Command,
	}
	if slices.Contains(required, "") {
		problems = append(problems, "registry id/title/generated_doc/workflow/evidence/precommit/command required")
	}
	problems = append(problems, validateExistingPath(root, reg.Workflow, "workflow")...)
	return problems
}

func validateExistingPath(root, rel, owner string) []string {
	if rel == "" || localFileExists(repoPath(root, rel)) {
		return nil
	}
	return []string{owner + " references missing path " + rel}
}
