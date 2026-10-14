package main

import "strings"

func validateForbiddenDocTokens(doc string, tokens []string) []problem {
	var problems []problem
	for _, token := range tokens {
		if strings.Contains(doc, token) {
			problems = append(problems, problem{Message: "generated doc contains stale state token " + token})
		}
	}
	return problems
}
