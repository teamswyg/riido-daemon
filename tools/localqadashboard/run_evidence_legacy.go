package main

import "encoding/json"

type legacyRunCandidates struct {
	Candidates []json.RawMessage `json:"closed_loop_candidates,omitempty"`
}

func (run *localRunEvidence) applyLegacyClosedLoopCandidates(data []byte) error {
	var legacy legacyRunCandidates
	if err := json.Unmarshal(data, &legacy); err != nil {
		return err
	}
	if len(legacy.Candidates) == 0 {
		return nil
	}
	filtered := make([]closedLoopCandidate, 0, len(run.Candidates))
	for idx, candidate := range run.Candidates {
		if !isLegacyProductLoopCandidate(candidate) || idx >= len(legacy.Candidates) {
			filtered = append(filtered, candidate)
			continue
		}
		var loop localRunLoopCandidate
		if err := json.Unmarshal(legacy.Candidates[idx], &loop); err != nil {
			return err
		}
		run.ClosedLoops = append(run.ClosedLoops, loop)
	}
	run.Candidates = filtered
	return nil
}

func isLegacyProductLoopCandidate(candidate closedLoopCandidate) bool {
	return candidate.Source == "" &&
		candidate.Trigger == "" &&
		candidate.Status == "" &&
		candidate.Summary == "" &&
		candidate.NextAction == ""
}
