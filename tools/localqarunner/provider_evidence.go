package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func applyProviderEvidence(root string, cfg config, evidence *runEvidence) error {
	data, err := os.ReadFile(outputPath(root, *cfg.providerEvidence))
	if err != nil {
		return fmt.Errorf("read provider evidence: %w", err)
	}
	var file providerRunEvidence
	if err := json.Unmarshal(data, &file); err != nil {
		return fmt.Errorf("parse provider evidence: %w", err)
	}
	evidence.ProviderStatus = file.Status
	evidence.ProviderSummary = providerSummaries(file.Providers)
	evidence.CoverageStatus = mergeCoverageStatus(evidence.CoverageStatus, file.Status)
	appendProviderRepairs(evidence, file.Providers)
	return nil
}

func providerSummaries(providers []providerRunProvider) []runProviderSummary {
	out := make([]runProviderSummary, 0, len(providers))
	for _, provider := range providers {
		out = append(out, runProviderSummary{
			ID:                provider.ID,
			Available:         provider.Available,
			Version:           provider.Version,
			IntegrationStatus: provider.IntegrationStatus,
			Observed:          provider.Observed,
			Repair:            providerRepair(provider),
		})
	}
	return out
}

func providerRepair(provider providerRunProvider) *runRepair {
	if provider.Repair == nil {
		return nil
	}
	repair := *provider.Repair
	repair.ProviderID = provider.ID
	return &repair
}

func mergeCoverageStatus(current, observed string) string {
	if current == "" {
		current = statusPassed
	}
	if current == statusFailed || observed == statusFailed {
		return statusFailed
	}
	if observed != "" && observed != statusPassed {
		return statusPartial
	}
	return current
}

func appendProviderRepairs(evidence *runEvidence, providers []providerRunProvider) {
	for _, provider := range providers {
		repair := providerRepair(provider)
		if repair == nil {
			continue
		}
		evidence.OpenRepairs = append(evidence.OpenRepairs, *repair)
	}
}
