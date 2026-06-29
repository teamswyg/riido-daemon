package main

import "fmt"

const (
	dailyEvidenceSweepID         = "daily-evidence-sweep"
	dailyEvidenceSweepName       = "Daily Evidence Sweep"
	dailyEvidenceSweepCommonName = "하루 한 번 증적 순회"
)

func dailyTriggerEvidence(cfg config) triggerEvidence {
	return triggerEvidence{
		ID:                       dailyEvidenceSweepID,
		Name:                     dailyEvidenceSweepName,
		CommonName:               dailyEvidenceSweepCommonName,
		Cadence:                  "daily",
		TimeLocal:                fmt.Sprintf("%02d:%02d", *cfg.hour, *cfg.minute),
		EntryPoint:               "go run ./tools/localqarunner",
		FreshnessWindow:          "24h by default; localqarunner -valid-for can override it",
		ClosedLoop:               "run probes, write evidence, render the dashboard, publish latest/timestamped artifacts, and expose stale rows after expires_at",
		RefreshesExpiredEvidence: true,
		Evidence: []string{
			*cfg.productEvidence,
			*cfg.coverageEvidence,
			".riido-local/evidence/local-qa-run.json",
		},
	}
}
