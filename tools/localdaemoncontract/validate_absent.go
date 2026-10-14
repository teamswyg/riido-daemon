package main

func validateAbsent(repo string, surfaces []AbsentSurface) ([]problem, []AbsentEvidence) {
	var problems []problem
	var evidence []AbsentEvidence
	for _, surface := range surfaces {
		if surface.Name == "" || len(surface.Scope) == 0 || len(surface.Tokens) == 0 {
			problems = append(problems, problem{Message: "invalid absent surface " + surface.Name})
			continue
		}
		nextProblems, nextEvidence := validateAbsentSurface(repo, surface)
		problems = append(problems, nextProblems...)
		evidence = append(evidence, nextEvidence...)
	}
	return problems, evidence
}

func validateAbsentSurface(repo string, surface AbsentSurface) ([]problem, []AbsentEvidence) {
	var problems []problem
	var evidence []AbsentEvidence
	for _, scope := range surface.Scope {
		for _, token := range surface.Tokens {
			hit, err := scopeContains(repoPath(repo, scope), token)
			ok := err == nil && !hit
			evidence = append(evidence, AbsentEvidence{Name: surface.Name, Scope: scope, Token: token, OK: ok})
			if err != nil {
				problems = append(problems, problem{Message: surface.Name + ": " + err.Error()})
			} else if hit {
				problems = append(problems, problem{Message: "forbidden token in " + scope + ": " + token})
			}
		}
	}
	return problems, evidence
}
