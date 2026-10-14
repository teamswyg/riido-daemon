package main

func validateManifest(m manifest) []string {
	var problems []string
	if m.SchemaVersion != "riido-context-map-docs.v1" {
		problems = append(problems, "unexpected schema_version")
	}
	if m.ID == "" || m.GeneratedDoc == "" || m.Workflow == "" || m.EvidenceArtifact == "" {
		problems = append(problems, "id, generated_doc, workflow, and evidence_artifact are required")
	}
	if len(m.FocusedSections) != 7 || len(m.Contexts) == 0 {
		problems = append(problems, "focused sections and contexts are required")
	}
	problems = append(problems, validateContexts(m.Contexts)...)
	problems = append(problems, validateFragments(m)...)
	return problems
}

func validateContexts(rows []contextRow) []string {
	var problems []string
	seen := map[string]bool{}
	for _, row := range rows {
		if row.ID == "" || row.Context == "" || row.Owner == "" {
			problems = append(problems, "context id, context, and owner are required")
		}
		if seen[row.ID] {
			problems = append(problems, "duplicate context id "+row.ID)
		}
		seen[row.ID] = true
	}
	return problems
}

func validateFragments(m manifest) []string {
	var problems []string
	if len(m.ACL.Rows) == 0 || len(m.Dependency.Diagram) == 0 || len(m.ChangeProcedure.SamePRUpdates) == 0 {
		problems = append(problems, "acl, dependency, and change fragments must not be empty")
	}
	problems = append(problems, validateFigmaSections(m.FigmaDaemon.Sections)...)
	problems = append(problems, validateFigmaSections(m.FigmaOnboarding.Sections)...)
	if len(m.SplitRepo.Rules) == 0 || len(m.SplitRepo.DaemonMustNotRedefine) == 0 {
		problems = append(problems, "split repo rules must not be empty")
	}
	return problems
}
