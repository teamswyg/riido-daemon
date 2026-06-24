package main

import (
	"fmt"
	"slices"
	"strings"
)

var (
	allowedDispositions = []string{"adopted", "triage_required", "deferred", "rejected"}
	allowedPriorities   = []string{"P0", "P1", "P2", "P3"}
)

func verifyDecision(decision decisionRecord) error {
	required := []string{
		decision.CandidateID, decision.Disposition, decision.Priority,
		decision.Owner, decision.NextLoop, decision.NextArtifact, decision.Reason,
	}
	for _, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("candidate decision fields must be complete")
		}
	}
	if !slices.Contains(allowedDispositions, decision.Disposition) {
		return fmt.Errorf("candidate %s has unknown disposition %s", decision.CandidateID, decision.Disposition)
	}
	if !slices.Contains(allowedPriorities, decision.Priority) {
		return fmt.Errorf("candidate %s has unknown priority %s", decision.CandidateID, decision.Priority)
	}
	if decisionNeedsReviewBy(decision) && decision.ReviewBy == "" {
		return fmt.Errorf("candidate %s needs review_by", decision.CandidateID)
	}
	return nil
}

func decisionNeedsReviewBy(decision decisionRecord) bool {
	return decision.Disposition == "triage_required" || decision.Disposition == "deferred"
}
