package main

func validateAbsent(repo string, surfaces []AbsentSurface, problems *[]problem) []AbsentCheck {
	results := make([]AbsentCheck, 0, len(surfaces))
	for _, surface := range surfaces {
		check := AbsentCheck{Name: surface.Name, Pass: true}
		if surface.Name == "" || len(surface.Scope) == 0 || len(surface.Tokens) == 0 {
			*problems = append(*problems, problem{"absent surfaces require name, scope, and tokens"})
		}
		for _, scope := range surface.Scope {
			scanAbsentPath(repo, repoPath(repo, scope), surface.Tokens, &check, problems)
		}
		check.Pass = len(check.Hits) == 0
		if !check.Pass {
			*problems = append(*problems, problem{"absent surface found: " + surface.Name})
		}
		results = append(results, check)
	}
	return results
}
