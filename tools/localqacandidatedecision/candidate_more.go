package main

import "slices"

func decisionsByID(decisions []decisionRecord) map[string]decisionRecord {
	out := map[string]decisionRecord{}
	for _, decision := range decisions {
		out[decision.CandidateID] = decision
	}
	return out
}

func orphanDecisionProblems(decisions []decisionRecord, candidates []closedLoopCandidate) []candidateProblem {
	candidateByID := map[string]bool{}
	for _, item := range candidates {
		candidateByID[item.ID] = true
	}
	var problems []candidateProblem
	for _, decision := range decisions {
		if decisionAllowsMissingCandidate(decision) {
			continue
		}
		if !candidateByID[decision.CandidateID] {
			problems = append(problems, orphanDecisionProblem(decision))
		}
	}
	return problems
}

func scopedDecisions(decisions []decisionRecord, scope string) []decisionRecord {
	var out []decisionRecord
	for _, decision := range decisions {
		if decisionMatchesCandidateScope(decision, scope) {
			out = append(out, decision)
		}
	}
	return out
}

func invalidArtifactProblem(candidate closedLoopCandidate, decision decisionRecord) (candidateProblem, bool) {
	if slices.Contains(candidate.RequiredNextArtifacts, decision.NextArtifact) {
		return candidateProblem{}, false
	}
	return candidateProblem{
		CandidateID:           candidate.ID,
		Reason:                "decision next_artifact is not required by candidate",
		RequiredNextArtifacts: candidate.RequiredNextArtifacts,
		DecisionNextArtifact:  decision.NextArtifact,
		RecommendedAction:     "Choose one next_artifact from required_next_artifacts.",
	}, true
}

func (g closedLoopEvidenceGraph) complete() bool {
	return g.Observation != "" &&
		g.Hypothesis != "" &&
		g.Change != "" &&
		g.Verifier != "" &&
		g.Evidence != "" &&
		g.Decision != "" &&
		g.NextLoop != ""
}
