package main

import (
	"fmt"
	"os"
	"strings"
)

func emitGitHubAnnotations(summary changedSummary) {
	for _, problem := range summary.ProblemDetails {
		fmt.Fprintln(os.Stderr, githubAnnotation(problem))
	}
}

func githubAnnotation(problem changedProblem) string {
	message := "claim " + problem.ClaimID + ": " + problem.Reason +
		". Update one bound evidence file: " + strings.Join(problem.RequiredEvidence, ", ")
	return "::error file=" + escapeCommand(firstChangedFile(problem)) +
		",title=Loop Registry Claim Binding::" + escapeCommand(message)
}

func escapeCommand(value string) string {
	value = strings.ReplaceAll(value, "%", "%25")
	value = strings.ReplaceAll(value, "\r", "%0D")
	value = strings.ReplaceAll(value, "\n", "%0A")
	return value
}
