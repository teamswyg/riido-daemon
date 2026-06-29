package main

func validateManifest(repo string, manifest Manifest) []problem {
	var problems []problem
	if manifest.SchemaVersion != "riido-semantic-change-bindings.v1" {
		problems = append(problems, problem{Message: "invalid schema_version"})
	}
	if manifest.ID == "" {
		problems = append(problems, problem{Message: "manifest id is required"})
	}
	for _, binding := range manifest.Bindings {
		problems = append(problems, validateBinding(repo, binding)...)
	}
	return problems
}

func validateBinding(repo string, binding Binding) []problem {
	var problems []problem
	if binding.ID == "" || binding.Claim == "" {
		problems = append(problems, problem{Message: "binding id and claim are required"})
	}
	problems = append(problems, validatePaths(repo, binding.ID, binding.Triggers)...)
	problems = append(problems, validatePaths(repo, binding.ID, binding.RequiredWithTriggers)...)
	problems = append(problems, validatePaths(repo, binding.ID, binding.GeneratedDocs)...)
	if len(binding.Verifiers) == 0 {
		problems = append(problems, problem{Message: binding.ID + " has no verifiers"})
	}
	return problems
}

func validatePaths(repo, bindingID string, paths []string) []problem {
	var problems []problem
	for _, path := range paths {
		if path == "" {
			problems = append(problems, problem{Message: bindingID + " contains empty path"})
			continue
		}
		if !pathExists(repo, path) {
			problems = append(problems, problem{Message: bindingID + " missing path " + path})
		}
	}
	return problems
}
