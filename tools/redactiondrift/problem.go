package main

import "strings"

func joinProblems(problems []string) string {
	var out strings.Builder
	for _, problem := range problems {
		out.WriteString("- ")
		out.WriteString(problem)
		out.WriteByte('\n')
	}
	return out.String()
}
