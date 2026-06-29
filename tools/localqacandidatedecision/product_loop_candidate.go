package main

import "strings"

func generatedProductLoopCandidateID(id string) bool {
	if strings.HasPrefix(id, "product-outcome-evidence-") {
		return true
	}
	return generatedPartialReductionCandidateID(id)
}

func generatedPartialReductionCandidateID(id string) bool {
	switch id {
	case "partial-reduction-open-candidate-debt",
		"partial-reduction-candidate-aging",
		"partial-reduction-local-qa-run-evidence":
		return true
	default:
		return false
	}
}
