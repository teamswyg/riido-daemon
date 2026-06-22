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
	evidence.CoverageStatus = mergeCoverageStatus(evidence.CoverageStatus, file.Status)
	appendProviderRepairs(evidence, file.Providers)
	return nil
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
		if provider.Repair == nil {
			continue
		}
		repair := *provider.Repair
		repair.ProviderID = provider.ID
		evidence.OpenRepairs = append(evidence.OpenRepairs, repair)
	}
}
