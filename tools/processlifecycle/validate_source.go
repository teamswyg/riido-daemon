package main

import (
	"fmt"
	"strings"
)

func validateSources(repo string, manifest Manifest) ([]problem, []SourceResult) {
	var problems []problem
	results := make([]SourceResult, 0, len(manifest.SourceChecks))
	for _, check := range manifest.SourceChecks {
		source, err := readSource(repo, check.File)
		result := SourceResult{Name: check.Name, File: check.File, Contains: check.Contains}
		result.Pass = err == nil && strings.Contains(source, check.Contains)
		if !result.Pass {
			problems = append(problems, problem{fmt.Sprintf("source drift: %s", check.Name)})
		}
		results = append(results, result)
	}
	return problems, results
}
