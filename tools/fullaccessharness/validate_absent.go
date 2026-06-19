package main

import "fmt"

func validateAbsent(repo string, surfaces []AbsentSurface) ([]problem, []AbsentEvidence) {
	var problems []problem
	evidence := make([]AbsentEvidence, 0, len(surfaces))
	for _, surface := range surfaces {
		ok := true
		for _, scope := range surface.Scope {
			for _, token := range surface.Tokens {
				found, err := scopeContains(repoPath(repo, scope), token)
				if err != nil {
					problems = append(problems, problem{Message: err.Error()})
					ok = false
					continue
				}
				if found {
					msg := fmt.Sprintf("%s: forbidden token %q in %s", surface.Name, token, scope)
					problems = append(problems, problem{Message: msg})
					ok = false
				}
			}
		}
		evidence = append(evidence, AbsentEvidence{Name: surface.Name, Scope: surface.Scope, OK: ok})
	}
	return problems, evidence
}
