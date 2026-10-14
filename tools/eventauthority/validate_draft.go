package main

func validateDraftFields(repo string, manifest Manifest) ([]FieldCheck, []problem) {
	fields, err := draftFields(repo, manifest)
	if err != nil {
		return nil, []problem{{err.Error()}}
	}
	var checks []FieldCheck
	var problems []problem
	for _, field := range manifest.DraftSuppliedFields {
		_, ok := fields[field]
		checks = append(checks, FieldCheck{Field: field, Expected: "present_in_draft", Actual: present(ok), Passed: ok})
		if !ok {
			problems = append(problems, problem{"draft missing supplied field: " + field})
		}
	}
	for _, field := range manifest.IngestorAssignedFields {
		_, ok := fields[field]
		checks = append(checks, FieldCheck{Field: field, Expected: "absent_from_draft", Actual: present(ok), Passed: !ok})
		if ok {
			problems = append(problems, problem{"ingestor assigned field exposed in draft: " + field})
		}
	}
	return checks, problems
}

func present(ok bool) string {
	if ok {
		return "present"
	}
	return "absent"
}
