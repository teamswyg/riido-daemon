package main

func evaluate(manifest Manifest, changed []string) ([]bindingResult, []problem) {
	changedSet := stringSet(changed)
	var results []bindingResult
	var problems []problem
	for _, binding := range manifest.Bindings {
		result := bindingResult{ID: binding.ID, Claim: binding.Claim}
		for _, trigger := range binding.Triggers {
			if changedSet[trigger] {
				result.Triggered = true
				result.ChangedTriggers = append(result.ChangedTriggers, trigger)
			}
		}
		if result.Triggered {
			result.MissingRequired = missingRequired(binding, changedSet)
			if len(result.MissingRequired) > 0 {
				problems = append(problems, problem{Message: result.ID + " missing required semantic peers: " + join(result.MissingRequired)})
			}
		}
		result.ChangedRequiredCount = countChanged(binding.RequiredWithTriggers, changedSet)
		results = append(results, result)
	}
	return results, problems
}

func missingRequired(binding Binding, changed map[string]bool) []string {
	var out []string
	for _, required := range binding.RequiredWithTriggers {
		if !changed[required] {
			out = append(out, required)
		}
	}
	return out
}

func countChanged(paths []string, changed map[string]bool) int {
	count := 0
	for _, path := range paths {
		if changed[path] {
			count++
		}
	}
	return count
}

func stringSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		out[value] = true
	}
	return out
}
