package main

import "time"

func figmaRefreshScenario(observed time.Time, validFor time.Duration, manifestPath, goldenPath string) scenario {
	proof := figmaRefreshProofAt(observed, validFor, manifestPath, goldenPath)
	status := statusPassed
	var repairEvidence *repair
	if proof.Err != "" {
		status = statusFailed
	} else if proof.Stale {
		status = statusPartial
		repairEvidence = &repair{
			Class:   "figma_refresh_required",
			Owner:   "local-qa",
			Mode:    "system-detected",
			Summary: "Figma golden evidence is older than the QA freshness window.",
		}
	}
	return scenario{
		ID:             "local.qa.figma_refresh_gate",
		Status:         status,
		Observed:       proof.Observed(),
		FailureSummary: proof.Err,
		Repair:         repairEvidence,
	}
}

type figmaRefreshProof struct {
	ManifestPath string
	GoldenPath   string
	CapturedAt   string
	AgeSeconds   int64
	ValidSeconds int64
	EntryCount   int
	ScreenCount  int
	Stale        bool
	Err          string
}

func (p figmaRefreshProof) Observed() map[string]any {
	return map[string]any{
		"mode":                 "system",
		"replaces_inferred_id": "figma-refresh",
		"manifest":             p.ManifestPath,
		"golden":               p.GoldenPath,
		"captured_at":          p.CapturedAt,
		"age_seconds":          p.AgeSeconds,
		"valid_for_seconds":    p.ValidSeconds,
		"entry_count":          p.EntryCount,
		"screen_count":         p.ScreenCount,
		"stale":                p.Stale,
		"error":                p.Err,
	}
}
