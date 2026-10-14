package main

func validateCore(core coreDoc) []string {
	var problems []string
	if core.ID == "" || core.Title == "" || core.GeneratedDoc == "" || core.Context == "" {
		problems = append(problems, "core id, title, generated_doc, and context are required")
	}
	if len(core.Responsibilities) == 0 || len(core.NonResponsibilities) == 0 {
		problems = append(problems, "core responsibilities and non-responsibilities are required")
	}
	if len(core.Invariants) != 9 {
		problems = append(problems, "nine core invariants are required")
	}
	for _, inv := range core.Invariants {
		if inv.Name == "" || inv.Summary == "" || len(inv.SourceChecks) == 0 {
			problems = append(problems, "invariant name, summary, and source checks are required")
		}
	}
	return problems
}

func validateInvariantChecks(m manifest) []string {
	known := map[string]bool{}
	for _, check := range m.SourceChecks {
		known[check.Name] = true
	}
	var problems []string
	for _, inv := range m.Core.Invariants {
		for _, name := range inv.SourceChecks {
			if !known[name] {
				problems = append(problems, "unknown invariant source check "+name)
			}
		}
	}
	return problems
}
