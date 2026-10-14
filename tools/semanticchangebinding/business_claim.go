package main

import "strings"

const businessClaimClass = "business"

func validateBusinessClaim(binding Binding) []problem {
	if binding.ClaimClass != businessClaimClass {
		return nil
	}
	paths := bindingSemanticPaths(binding)
	checks := []struct {
		ok      bool
		message string
	}{
		{hasCodePath(paths), " has no code path"},
		{hasTestPath(paths), " has no test path"},
		{hasDocPath(paths), " has no documentation path"},
		{len(binding.GeneratedDocs) > 0, " has no generated docs"},
		{len(binding.EvidenceIDs) > 0, " has no evidence ids"},
	}
	var problems []problem
	for _, check := range checks {
		if !check.ok {
			problems = append(problems, problem{Message: binding.ID + check.message})
		}
	}
	return problems
}

func bindingSemanticPaths(binding Binding) []string {
	var paths []string
	paths = append(paths, binding.Triggers...)
	paths = append(paths, binding.RequiredWithTriggers...)
	paths = append(paths, binding.GeneratedDocs...)
	return paths
}

func hasCodePath(paths []string) bool {
	for _, path := range paths {
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			return strings.HasPrefix(path, "internal/") || strings.HasPrefix(path, "tools/")
		}
	}
	return false
}

func hasTestPath(paths []string) bool {
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") || strings.HasPrefix(path, ".github/workflows/") {
			return true
		}
	}
	return false
}

func hasDocPath(paths []string) bool {
	for _, path := range paths {
		if strings.HasPrefix(path, "docs/") {
			return true
		}
	}
	return false
}
