package main

import "time"

func expiredCoverageRows(coverage *runCoverage) int {
	if coverage == nil {
		return 0
	}
	now := time.Now().UTC()
	var count int
	for _, row := range coverage.Rows {
		if coverageRowExpired(row, now) {
			count++
		}
	}
	return count
}

func coverageRowExpired(row runCoverageRow, now time.Time) bool {
	if row.ExpiresAt == "" {
		return false
	}
	expires, err := time.Parse(time.RFC3339, row.ExpiresAt)
	if err != nil {
		return false
	}
	return !now.Before(expires)
}

func appendExpiredCoverageBlocker(blockers []runDeploymentGateBlocker, count int) []runDeploymentGateBlocker {
	if count == 0 {
		return blockers
	}
	return append(blockers, runDeploymentGateBlocker{
		Code:    "coverage_evidence_expired",
		Summary: "local QA coverage contains expired evidence rows",
		Count:   count,
	})
}
