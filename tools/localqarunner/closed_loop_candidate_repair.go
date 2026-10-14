package main

func appendRepairCandidates(out []closedLoopCandidate, seen map[string]bool,
	evidence runEvidence,
) []closedLoopCandidate {
	for _, repair := range evidence.OpenRepairs {
		id := "open-repair." + stableID(repair.ProviderID+"."+repair.Class)
		next := repair.SuggestedCommand
		if next == "" {
			next = "Close the repair manually, or promote it to a runner-backed verifier."
		}
		out = appendCandidate(out, seen, closedLoopCandidate{
			ID:         id,
			Source:     "open_repairs",
			Trigger:    "repair_required",
			Summary:    repair.Summary,
			Evidence:   evidence.Artifacts.ProviderEvidence,
			NextAction: next,
		})
	}
	return out
}
