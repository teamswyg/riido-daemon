package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func validateAbsent(repo string, surfaces []AbsentSurface) ([]problem, []AbsentCheck) {
	var problems []problem
	checks := make([]AbsentCheck, 0, len(surfaces))
	for _, surface := range surfaces {
		check := absentCheck(repo, surface, &problems)
		if !check.Pass {
			problems = append(problems, problem{fmt.Sprintf("unexpected surface: %s", surface.Name)})
		}
		checks = append(checks, check)
	}
	return problems, checks
}

func absentCheck(repo string, surface AbsentSurface, problems *[]problem) AbsentCheck {
	check := AbsentCheck{Name: surface.Name, Scope: surface.Scope, Tokens: surface.Tokens, Hits: []string{}, Pass: true}
	for _, scope := range surface.Scope {
		scanScope(repo, scope, surface.Tokens, &check, problems)
	}
	check.Pass = len(check.Hits) == 0
	return check
}

func scanScope(repo, scope string, tokens []string, check *AbsentCheck, problems *[]problem) {
	root, err := cleanRepoPath(repo, scope)
	if err != nil {
		*problems = append(*problems, problem{err.Error()})
		return
	}
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		scanFile(repo, path, tokens, check)
		return nil
	})
	if err != nil {
		*problems = append(*problems, problem{err.Error()})
	}
}

func scanFile(repo, path string, tokens []string, check *AbsentCheck) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	source := string(data)
	for _, token := range tokens {
		if strings.Contains(source, token) {
			rel, err := filepath.Rel(repo, path)
			if err != nil {
				rel = path
			}
			check.Hits = append(check.Hits, fmt.Sprintf("%s:%s", rel, token))
		}
	}
}
