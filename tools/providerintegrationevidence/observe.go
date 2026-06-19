package main

import (
	"fmt"
	"time"
)

func observeProviders(root string, m manifest, runIntegration bool) (evidenceFile, error) {
	file := evidenceFile{
		SchemaVersion: "riido-provider-real-cli-observation-result.v1",
		ID:            m.ID,
		ObservedAt:    time.Now().UTC().Format(time.RFC3339),
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
	status := "skipped"
	for _, provider := range providers {
		switch provider.IntegrationStatus {
		case "failed":
			return "failed"
		case "passed":
			status = "passed"
		case "observed":
			if status == "skipped" {
				status = "observed"
			}
		}
	}
	return status
}
