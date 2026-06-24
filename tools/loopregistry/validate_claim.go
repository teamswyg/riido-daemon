package main

func validateClaims(root string, claims []businessClaim) []string {
	seen := map[string]bool{}
	var problems []string
	for _, claim := range claims {
		problems = append(problems, validateClaim(root, claim, seen)...)
	}
	return problems
}

func validateClaim(root string, claim businessClaim, seen map[string]bool) []string {
	var problems []string
	if claim.ID == "" || claim.Text == "" {
		return []string{"business claim id/text are required"}
	}
	if seen[claim.ID] {
		problems = append(problems, "duplicate business claim "+claim.ID)
	}
	seen[claim.ID] = true
	problems = append(problems, requireClaimList(claim.ID, "files", claim.Files)...)
	problems = append(problems, requireClaimList(claim.ID, "docs", claim.Docs)...)
	problems = append(problems, requireClaimList(claim.ID, "evidence", claim.Evidence)...)
	for _, rel := range append(claim.Files, claim.Docs...) {
		problems = append(problems, validateExistingPath(root, rel, claim.ID)...)
	}
	problems = append(problems, validateChecks(root, claim.ID, claim.Verifiers, "verifier")...)
	problems = append(problems, validateChecks(root, claim.ID, claim.Contracts, "contract")...)
	return problems
}

func requireClaimList(claimID, field string, values []string) []string {
	if len(values) == 0 {
		return []string{claimID + " requires " + field}
	}
	return nil
}
