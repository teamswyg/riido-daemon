package main

import (
	"fmt"
	"path/filepath"
	"time"
)

func run(cfg config) (string, error) {
	clientRoot, err := filepath.Abs(*cfg.clientRoot)
	if err != nil {
		return statusFailed, fmt.Errorf("resolve client root: %w", err)
	}
	hydrateConfigFromStorage(cfg)
	observed := time.Now().UTC()
	evidence := evidenceFile{
		SchemaVersion: "riido-product-acceptance.v1",
		ID:            "ai-agent-product-acceptance",
		ObservedAt:    observed.Format(time.RFC3339),
		ExpiresAt:     observed.Add(*cfg.validFor).Format(time.RFC3339),
		Scenarios:     buildScenarios(clientRoot, cfg),
	}
	evidence.Status = summarize(evidence.Scenarios)
	if err := writeEvidence(*cfg.evidenceOut, evidence); err != nil {
		return statusFailed, err
	}
	if err := writeContractLab(*cfg.labOut, evidence); err != nil {
		return statusFailed, err
	}
	return evidence.Status, nil
}

func summarize(scenarios []scenario) string {
	status := statusPassed
	for _, scenario := range scenarios {
		if scenario.Status == statusFailed {
			return statusFailed
		}
		if scenario.Status == statusSkipped {
			status = statusPartial
		}
	}
	return status
}

func buildScenarios(clientRoot string, cfg config) []scenario {
	out := []scenario{clientReadOnlyScenario(clientRoot)}
	out = append(out, contractAPIScenarios(cfg)...)
	out = append(out, figmaIntentScenarios(*cfg.figmaManifest, *cfg.figmaGolden, *cfg.screenshots)...)
	out = append(out, contractUIScenario(*cfg.labOut))
	return out
}
