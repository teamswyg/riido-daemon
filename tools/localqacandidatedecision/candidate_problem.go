package main

import "strings"

type candidateProblem struct {
	CandidateID           string
	Reason                string
	RequiredNextArtifacts []string
	DecisionNextArtifact  string
	RecommendedAction     string
}

type candidateDecisionError struct {
	Problems []candidateProblem
}

func (e candidateDecisionError) Error() string {
	var out []string
	for _, problem := range e.Problems {
		out = append(out, problem.summary())
	}
	return strings.Join(out, "; ")
}

func (p candidateProblem) summary() string {
	return "candidate " + p.CandidateID + ": " + p.Reason
}
