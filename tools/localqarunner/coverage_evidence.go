package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type runCoverage struct {
	Summary runCoverageSummary `json:"summary"`
	Rows    []runCoverageRow   `json:"rows"`
}

type runCoverageSummary struct {
	Total       int `json:"total"`
	Passed      int `json:"passed"`
	Skipped     int `json:"skipped"`
	NotVerified int `json:"not_verified"`
	Failed      int `json:"failed"`
}

type runCoverageRow struct {
	ID      string     `json:"id"`
	Title   string     `json:"title"`
	Tier    string     `json:"tier"`
	Surface string     `json:"surface"`
	Status  string     `json:"status"`
	Repair  *runRepair `json:"repair,omitempty"`
	Detail  string     `json:"detail,omitempty"`
}

func applyCoverageEvidence(root string, cfg config, evidence *runEvidence) error {
	data, err := os.ReadFile(outputPath(root, *cfg.coverageEvidence))
	if err != nil {
		return fmt.Errorf("read coverage evidence: %w", err)
	}
	var coverage runCoverage
	if err := json.Unmarshal(data, &coverage); err != nil {
		return fmt.Errorf("parse coverage evidence: %w", err)
	}
	evidence.Coverage = &coverage
	evidence.CoverageStatus = coverageStatus(coverage.Summary)
	return nil
}

func coverageStatus(summary runCoverageSummary) string {
	if summary.Failed > 0 {
		return statusFailed
	}
	if summary.Total == 0 || summary.Passed != summary.Total {
		return statusPartial
	}
	return statusPassed
}
