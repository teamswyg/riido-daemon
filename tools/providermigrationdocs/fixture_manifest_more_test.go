package main

func validProviderMigrationManifest() manifest {
	return manifest{
		SchemaVersion:    schemaVersion,
		ID:               "provider-migration",
		GeneratedDoc:     "root.md",
		Workflow:         "workflow.yml",
		EvidenceArtifact: "provider-migration",
		Pages: []page{{
			ID: "p1", Title: "One", GeneratedDoc: "p1.md",
			Artifacts: []string{"missing.md"},
		}},
		Assertions: []string{"docs are generated"},
	}
}
