package main

import "time"

func previousCandidateMap(candidates []closedLoopCandidate) map[string]closedLoopCandidate {
	out := make(map[string]closedLoopCandidate, len(candidates))
	for _, candidate := range candidates {
		out[candidate.ID] = candidate
	}
	return out
}

func candidateFirstSeen(now time.Time, current, previous closedLoopCandidate) time.Time {
	if previous.FirstObservedAt != "" {
		return parseCandidateTime(previous.FirstObservedAt)
	}
	if current.FirstObservedAt != "" {
		return parseCandidateTime(current.FirstObservedAt)
	}
	return now
}
