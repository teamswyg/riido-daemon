package main

func staticEvidenceGapCandidates(
	manualPresent bool,
	capturePresent bool,
	captureUploadCovered bool,
) []evidenceGapCandidate {
	out := []evidenceGapCandidate{}
	if !manualPresent {
		out = append(out, staticGap(
			"manual-evidence-file",
			"Manual QA state is browser-local until exported.",
			"Create .riido-local/evidence/manual-qa-evidence.json before S3 upload.",
		))
	}
	if capturePresent && !captureUploadCovered {
		out = append(out, staticGap(
			"contract-lab-capture-upload",
			"Feature UI capture exists outside the current product screenshot upload dir.",
			"Upload screenshots/contract-lab or move generated captures under the uploaded screenshot root.",
		))
	}
	out = append(out,
		staticGap(
			"browser-interaction-runner",
			"DSL declares interactions, but localproductacceptance does not replay them by itself.",
			"Add a small browser QA runner that emits interaction JSON and PNG.",
		),
		staticGap(
			"runtime-detail-golden",
			"Runtime detail has Figma intent evidence but no visual golden screenshot.",
			"Capture node 1179:27360 as a golden reference.",
		),
	)
	return out
}

func staticGap(id, reason, next string) evidenceGapCandidate {
	return closedLoopCandidate(id, "", "known_gap", reason, next)
}
