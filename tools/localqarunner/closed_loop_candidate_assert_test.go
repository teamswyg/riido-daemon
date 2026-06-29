package main

import "testing"

func assertCandidate(t *testing.T, rows []closedLoopCandidate, id string) {
	t.Helper()
	for _, row := range rows {
		if row.ID == id && row.Status == "candidate" && row.StaleAfterHours > 0 &&
			row.FirstObservedAt != "" && row.Graph.Observation != "" {
			return
		}
	}
	t.Fatalf("candidate %q missing from %+v", id, rows)
}
