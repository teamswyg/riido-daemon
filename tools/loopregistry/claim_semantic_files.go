package main

func claimsByID(claims []businessClaim) map[string]businessClaim {
	out := map[string]businessClaim{}
	for _, claim := range claims {
		out[claim.ID] = claim
	}
	return out
}

func claimEvidenceFiles(claim businessClaim) []string {
	out := append([]string{}, claim.Files...)
	out = append(out, claim.Docs...)
	for _, check := range append(claim.Verifiers, claim.Contracts...) {
		out = append(out, check.File)
	}
	return uniqueStrings(out)
}

func mergeClaimEvidence(previous, current businessClaim) []string {
	return uniqueStrings(append(claimEvidenceFiles(previous), claimEvidenceFiles(current)...))
}

func claimSemanticProblem(
	id string,
	required []string,
	changed []string,
) changedProblem {
	if intersects(changed, required) {
		return changedProblem{}
	}
	return changedProblem{
		ClaimID:          id,
		Reason:           "business claim changed without bound code/doc/test evidence",
		ChangedFiles:     intersection(changed, registryFiles()),
		RequiredEvidence: required,
	}
}

func compactClaimProblems(items []changedProblem) []changedProblem {
	var out []changedProblem
	for _, item := range items {
		if item.ClaimID != "" {
			out = append(out, item)
		}
	}
	return out
}
