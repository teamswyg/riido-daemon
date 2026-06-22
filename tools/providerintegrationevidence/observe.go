package main

import (
	"fmt"
	"runtime"
	"time"
)

func observeProviders(root string, m manifest, runIntegration bool, validFor time.Duration) (evidenceFile, error) {
	now := time.Now().UTC()
	file := evidenceFile{
		SchemaVersion:   "riido-provider-real-cli-observation-result.v1",
		ID:              m.ID,
		ObservedAt:      now.Format(time.RFC3339),
		ExpiresAt:       now.Add(validFor).Format(time.RFC3339),
		FreshForSeconds: int64(validFor.Seconds()),
		RunIntegration:  runIntegration,
		Platform: evidencePlatform{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		},
	}
	var failed []string
	for _, provider := range m.Providers {
		ev := observeProvider(root, provider, runIntegration)
		if ev.IntegrationStatus == "failed" {
			failed = append(failed, provider.ID)
		}
		file.Providers = append(file.Providers, ev)
	}
	file.Status = aggregateStatus(file.Providers)
	if len(failed) > 0 {
		return file, fmt.Errorf("provider integration failed: %v", failed)
	}
	return file, nil
}

func aggregateStatus(providers []providerEvidence) string {
	var observed, passed, skipped int
	for _, provider := range providers {
		switch provider.IntegrationStatus {
		case "failed":
			return "failed"
		case "passed":
			passed++
		case "observed":
			observed++
		case "skipped":
			skipped++
		}
	}
	if passed > 0 && skipped > 0 {
		return "partial"
	}
	if passed > 0 {
		return "passed"
	}
	if observed > 0 {
		return "observed"
	}
	return "skipped"
}
