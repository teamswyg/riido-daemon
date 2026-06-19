package main

func validateManifest(m manifest) []problem {
	var problems []problem
	if m.SchemaVersion != "riido-module-decomposition.v1" {
		problems = append(problems, problem{Message: "unexpected schema_version"})
	}
	if m.ID == "" || m.Title == "" || m.GeneratedDoc == "" || m.Workflow == "" {
		problems = append(problems, problem{Message: "id, title, generated_doc, and workflow are required"})
	}
	if m.ModulePath == "" || m.BinaryPackage == "" {
		problems = append(problems, problem{Message: "module_path and binary_package are required"})
	}
	return append(problems, validateCollections(m)...)
}

func validateCollections(m manifest) []problem {
	var problems []problem
	if len(m.DetailDocs) != 5 || len(m.PackageRoles) == 0 || len(m.ImportRules) == 0 {
		problems = append(problems, problem{Message: "detail docs, package roles, and import rules are required"})
	}
	if len(m.Decisions) == 0 || len(m.Ports) == 0 || len(m.Assertions) == 0 {
		problems = append(problems, problem{Message: "decisions, ports, and assertions are required"})
	}
	for _, doc := range m.DetailDocs {
		if doc.Title == "" || doc.Path == "" {
			problems = append(problems, problem{Message: "detail docs require title and path"})
		}
	}
	return problems
}
