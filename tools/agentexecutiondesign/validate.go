package main

func validateModel(m model) []string {
	var problems []string
	if m.Manifest.SchemaVersion != "riido-agent-execution-design-docs.v1" {
		problems = append(problems, "unexpected schema_version")
	}
	if m.Manifest.ID == "" || m.Manifest.GeneratedDoc == "" || m.Manifest.Workflow == "" {
		problems = append(problems, "id, generated_doc, and workflow are required")
	}
	if m.Manifest.EvidenceManifest == "" || m.Manifest.AssignmentFSMDoc == "" {
		problems = append(problems, "evidence_manifest and assignment_fsm_doc are required")
	}
	if len(m.Overview.SharedShape) == 0 || len(m.Overview.FocusedFiles) == 0 {
		problems = append(problems, "overview shared_shape and focused_files are required")
	}
	if len(m.Risk.Problems) == 0 || len(m.Execution.IdentityFields) == 0 {
		problems = append(problems, "risk problems and execution identity fields are required")
	}
	if len(m.Items) == 0 || len(m.Boundaries) == 0 {
		problems = append(problems, "evidence items and remaining boundaries are required")
	}
	return problems
}
