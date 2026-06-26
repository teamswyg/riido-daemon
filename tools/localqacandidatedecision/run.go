package main

func run(opt options) error {
	root, err := findRepoRoot(opt.Repo)
	if err != nil {
		return err
	}
	m, err := loadManifest(repoPath(root, opt.Manifest))
	if err != nil {
		return err
	}
	result, err := verifyAll(root, m)
	if err != nil {
		return err
	}
	if err := requireCandidateInput(opt); err != nil {
		return err
	}
	if opt.CandidateIn != "" {
		candidateResult, err := verifyCandidateDecisions(root, m, opt.CandidateIn, opt.CandidateScope)
		if err != nil {
			if opt.GitHubAnnotations {
				emitCandidateAnnotations(err)
			}
			return err
		}
		result.CandidateScope = candidateResult.CandidateScope
		result.CandidateCount = candidateResult.CandidateCount
		result.DecisionIDs = candidateResult.DecisionIDs
		result.DecisionArtifacts = candidateResult.DecisionArtifacts
	}
	if err := maybeDoc(root, m.GeneratedDoc, renderDoc(m, result), opt.WriteDoc, opt.CheckDoc); err != nil {
		return err
	}
	if opt.EvidenceOut != "" {
		return writeJSON(opt.EvidenceOut, newEvidence(m, result))
	}
	return nil
}

func requireCandidateInput(opt options) error {
	if opt.CandidateIn != "" || (!opt.CheckDoc && opt.EvidenceOut == "") {
		return nil
	}
	return errMissingCandidateInput
}
