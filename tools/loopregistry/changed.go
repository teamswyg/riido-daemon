package main

func changedCheck(
	root string,
	reg registry,
	path string,
	previousManifest string,
) changedSummary {
	changed, err := readChangedFiles(root, path)
	if err != nil {
		return changedSummary{Problems: []string{err.Error()}}
	}
	out := changedSummary{InputCount: len(changed)}
	for _, claim := range reg.BusinessClaims {
		if !intersects(changed, claim.Files) {
			continue
		}
		out.MatchedClaims = append(out.MatchedClaims, claim.ID)
		for _, detail := range validateClaimChange(claim, changed) {
			out.Problems = append(out.Problems, detail.summary())
			out.ProblemDetails = append(out.ProblemDetails, detail)
		}
	}
	for _, detail := range changedClaimProblems(root, previousManifest, reg, changed) {
		out.MatchedClaims = append(out.MatchedClaims, detail.ClaimID)
		out.Problems = append(out.Problems, detail.summary())
		out.ProblemDetails = append(out.ProblemDetails, detail)
	}
	out.MatchedClaims = uniqueStrings(out.MatchedClaims)
	out.MatchedClaimCount = len(out.MatchedClaims)
	return out
}

func claimDocs(claim businessClaim) []string {
	out := append([]string{}, claim.Docs...)
	for _, check := range append(claim.Verifiers, claim.Contracts...) {
		out = append(out, check.File)
	}
	return out
}

func registryFiles() []string {
	return []string{defaultManifest, "docs/30-architecture/loop-registry.md"}
}
