package main

func validateClaimChange(claim businessClaim, changed []string) []changedProblem {
	if !intersects(changed, claim.Files) {
		return nil
	}
	required := uniqueStrings(append(claimDocs(claim), registryFiles()...))
	if intersects(changed, required) {
		return nil
	}
	return []changedProblem{{
		ClaimID:          claim.ID,
		Reason:           "runtime files changed without bound doc/verifier/registry evidence",
		ChangedFiles:     intersection(changed, claim.Files),
		RequiredEvidence: required,
	}}
}
