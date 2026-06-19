package main

func validateInputs(m manifest, c contract) []string {
	var problems []string
	if m.SchemaVersion != "riido-store-distribution-docs.v1" {
		problems = append(problems, "unexpected schema_version")
	}
	if m.ID == "" || m.GeneratedDoc == "" || m.StoreContract == "" {
		problems = append(problems, "id, generated_doc, and store_contract are required")
	}
	if len(m.CompatibilityMarkers) == 0 || len(m.Decisions) == 0 {
		problems = append(problems, "compatibility_markers and decisions are required")
	}
	if len(m.DaemonChanges) == 0 || len(m.ServerChanges) == 0 {
		problems = append(problems, "daemon_changes and server_changes are required")
	}
	if c.Product == "" || len(c.Channels) == 0 {
		problems = append(problems, "store contract product and channels are required")
	}
	return problems
}
