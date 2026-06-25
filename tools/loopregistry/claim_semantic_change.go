package main

func changedClaimProblems(
	root string,
	previousManifest string,
	current registry,
	changed []string,
) []changedProblem {
	if previousManifest == "" {
		return nil
	}
	previous, err := loadRegistry(repoPath(root, previousManifest))
	if err != nil {
		return []changedProblem{{
			ClaimID: defaultManifest,
			Reason:  "read previous loop registry manifest: " + err.Error(),
		}}
	}
	return claimSemanticProblems(previous.BusinessClaims, current.BusinessClaims, changed)
}

func claimSemanticProblems(previous, current []businessClaim, changed []string) []changedProblem {
	prevByID := claimsByID(previous)
	currByID := claimsByID(current)
	var out []changedProblem
	for id, claim := range currByID {
		before, ok := prevByID[id]
		if ok && sameClaimSemantics(before, claim) {
			continue
		}
		out = append(out, claimSemanticProblem(id, mergeClaimEvidence(before, claim), changed))
	}
	for id, claim := range prevByID {
		if _, ok := currByID[id]; !ok {
			out = append(out, claimSemanticProblem(id, claimEvidenceFiles(claim), changed))
		}
	}
	return compactClaimProblems(out)
}
