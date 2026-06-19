package main

func validateManifest(m manifest) []problem {
	var problems []problem
	if m.SchemaVersion != "riido-figma-ai-agent-daemon-boundary.v1" {
		problems = append(problems, problem{Message: "unexpected schema_version"})
	}
	if m.ID == "" || m.RiidoTask == "" || m.HumanDoc == "" {
		problems = append(problems, problem{Message: "manifest id, riido_task, and human_doc are required"})
	}
	if m.BoundaryPolicy.Summary == "" || m.BoundaryPolicy.TopDown == "" || m.BoundaryPolicy.BottomUp == "" {
		problems = append(problems, problem{Message: "boundary policy must be complete"})
	}
	if len(m.Entries) == 0 {
		problems = append(problems, problem{Message: "boundary entries must not be empty"})
	}
	for _, entry := range m.Entries {
		problems = append(problems, validateEntry(entry)...)
	}
	return problems
}

func validateEntry(entry boundaryEntry) []problem {
	if entry.NodeID == "" || entry.Name == "" || entry.DaemonScope == "" {
		return []problem{{Message: "entry node_id, name, and daemon_scope are required"}}
	}
	if len(entry.UpstreamOwner) == 0 || entry.DaemonConsumedFacts == nil || len(entry.ClientOwnedFacts) == 0 {
		return []problem{{Message: "entry ownership and fact separation are required"}}
	}
	return nil
}
