package main

import (
	"fmt"
	"strings"
)

func validateForbidden(repo string, manifest Manifest) ([]problem, []ForbiddenCheck) {
	var problems []problem
	checks := make([]ForbiddenCheck, 0, len(manifest.ForbiddenFields))
	sources := loadForbiddenSources(repo, manifest.ForbiddenScope, &problems)
	for _, field := range manifest.ForbiddenFields {
		check := forbiddenCheck(manifest.ForbiddenScope, sources, field)
		if !check.Pass {
			problems = append(problems, problem{fmt.Sprintf("forbidden field leak: %s", field.Field)})
		}
		checks = append(checks, check)
	}
	return problems, checks
}

func loadForbiddenSources(repo string, scope []string, problems *[]problem) map[string]string {
	sources := make(map[string]string, len(scope))
	for _, path := range scope {
		source, err := readSource(repo, path)
		if err != nil {
			*problems = append(*problems, problem{err.Error()})
			continue
		}
		sources[path] = source
	}
	return sources
}

func forbiddenCheck(scope []string, sources map[string]string, field ForbiddenField) ForbiddenCheck {
	check := ForbiddenCheck{
		Field: field.Field, Scope: scope, Tokens: field.Tokens,
		Hits: []string{}, Pass: true,
	}
	for path, source := range sources {
		for _, token := range field.Tokens {
			if strings.Contains(source, token) {
				check.Hits = append(check.Hits, fmt.Sprintf("%s:%s", path, token))
			}
		}
	}
	check.Pass = len(check.Hits) == 0
	return check
}
