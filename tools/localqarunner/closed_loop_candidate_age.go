package main

import "time"

func annotateCandidateAge(observed string, candidates []closedLoopCandidate,
	previous []closedLoopCandidate,
) []closedLoopCandidate {
	now := parseCandidateTime(observed)
	prior := previousCandidateMap(previous)
	for i := range candidates {
		first := candidateFirstSeen(now, candidates[i], prior[candidates[i].ID])
		candidates[i].FirstObservedAt = first.Format(time.RFC3339)
		candidates[i].LastObservedAt = now.Format(time.RFC3339)
		candidates[i].AgeHours = int(now.Sub(first).Hours())
		candidates[i].StaleAt = first.Add(candidateStaleAfterHours * time.Hour).Format(time.RFC3339)
		candidates[i].Stale = !now.Before(first.Add(candidateStaleAfterHours * time.Hour))
		if candidates[i].Stale {
			candidates[i].Status = "stale"
		}
	}
	return candidates
}

func parseCandidateTime(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Now().UTC()
	}
	return t.UTC()
}
