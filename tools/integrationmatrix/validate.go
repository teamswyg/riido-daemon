package main

import "fmt"

func validateManifest(m manifest) []problem {
	var problems []problem
	if m.SchemaVersion != "riido-integration-matrix-docs.v1" {
		problems = append(problems, problem{Message: "unexpected schema_version"})
	}
	if m.ID == "" || m.Title == "" || m.GeneratedDoc == "" || m.Workflow == "" {
		problems = append(problems, problem{Message: "id, title, generated_doc, and workflow are required"})
	}
	if m.ProviderValidationManifest == "" || m.RealCLIObservationManifest == "" {
		problems = append(problems, problem{Message: "provider manifests are required"})
	}
	return append(problems, validateCollections(m)...)
}

func validateCollections(m manifest) []problem {
	var problems []problem
	if len(m.DetailDocs) != 5 || len(m.GatePolicy) == 0 || len(m.Assertions) == 0 {
		problems = append(problems, problem{Message: "detail docs, gate policy, and assertions are required"})
	}
	if len(m.ProviderValidation.Providers) == 0 || len(m.InstructionProbe.Providers) == 0 {
		problems = append(problems, problem{Message: "provider validation rows and instruction probes are required"})
	}
	for _, doc := range m.DetailDocs {
		if doc.Title == "" || doc.Path == "" {
			problems = append(problems, problem{Message: "detail docs require title and path"})
		}
	}
	return append(problems, validateProviderRows(m)...)
}

func validateProviderRows(m manifest) []problem {
	var problems []problem
	for _, p := range m.ProviderValidation.Providers {
		if p.Provider == "" || p.DisplayName == "" || p.OptInIntegration == "" {
			problems = append(problems, problem{Message: fmt.Sprintf("invalid provider row: %+v", p)})
		}
	}
	return problems
}
