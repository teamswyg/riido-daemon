package main

func providerRepairCandidateGraph(candidate runLoopCandidate) runLoopCandidateGraph {
	return runLoopCandidateGraph{
		Observation: candidate.Reason,
		Hypothesis:  "A focused provider repair loop can remove this harness failure from local QA.",
		Change:      candidate.NextEvidence,
		Verifier:    "local.qa.provider_repair_candidates",
		Evidence:    candidate.ID,
		Decision:    "Keep the candidate open until repair evidence and a verifier replace the manual gap.",
		NextLoop:    "promote-" + candidate.SourceScenario + "-" + candidate.ID,
	}
}
