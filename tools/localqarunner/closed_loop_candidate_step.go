package main

func appendStepCandidates(out []closedLoopCandidate, seen map[string]bool,
	evidence runEvidence,
) []closedLoopCandidate {
	for _, step := range evidence.Steps {
		if step.Status == statusPassed {
			continue
		}
		summary := step.OutputTail
		if summary == "" {
			summary = "local QA step did not pass"
		}
		out = appendCandidate(out, seen, closedLoopCandidate{
			ID:         "harness-step." + stableID(step.ID),
			Source:     "steps",
			Trigger:    "harness_step_not_passed",
			Summary:    summary,
			Evidence:   step.Command,
			NextAction: "Promote this failing step into a verifier or an explicit owner repair.",
		})
	}
	return out
}
