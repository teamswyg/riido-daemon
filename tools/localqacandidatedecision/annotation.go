package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func emitCandidateAnnotations(err error) {
	var decisionErr candidateDecisionError
	if !errors.As(err, &decisionErr) {
		return
	}
	for _, problem := range decisionErr.Problems {
		fmt.Fprintln(os.Stderr, githubAnnotation(problem))
	}
}

func githubAnnotation(problem candidateProblem) string {
	return "::error file=" + defaultManifest +
		",title=Local QA Candidate Decision::" +
		escapeCommand(candidateAnnotationMessage(problem))
}

func candidateAnnotationMessage(problem candidateProblem) string {
	msg := problem.summary()
	if len(problem.RequiredNextArtifacts) > 0 {
		msg += ". required_next_artifacts=" + strings.Join(problem.RequiredNextArtifacts, ",")
	}
	if problem.DecisionNextArtifact != "" {
		msg += ". decision_next_artifact=" + problem.DecisionNextArtifact
	}
	if problem.RecommendedAction != "" {
		msg += ". " + problem.RecommendedAction
	}
	return msg
}
