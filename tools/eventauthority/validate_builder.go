package main

func validateBuilderFields(repo string, manifest Manifest) ([]FieldCheck, []problem) {
	fields, err := builderFields(repo, manifest)
	if err != nil {
		return nil, []problem{{err.Error()}}
	}
	var checks []FieldCheck
	var problems []problem
	for _, field := range append(manifest.DraftSuppliedFields, manifest.IngestorAssignedFields...) {
		_, ok := fields[field]
		checks = append(checks, FieldCheck{Field: field, Expected: "assigned_in_builder", Actual: present(ok), Passed: ok})
		if !ok {
			problems = append(problems, problem{"builder missing CanonicalEvent field: " + field})
		}
	}
	return checks, problems
}
