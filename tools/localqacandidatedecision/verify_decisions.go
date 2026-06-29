package main

import "fmt"

func verifyDecisions(m manifest) error {
	if len(m.Decisions) == 0 || len(m.Assertions) == 0 {
		return fmt.Errorf("decisions and assertions are required")
	}
	seen := map[string]bool{}
	for _, decision := range m.Decisions {
		if seen[decision.CandidateID] {
			return fmt.Errorf("duplicate decision for candidate %s", decision.CandidateID)
		}
		seen[decision.CandidateID] = true
		if err := verifyDecision(decision); err != nil {
			return err
		}
	}
	return nil
}
