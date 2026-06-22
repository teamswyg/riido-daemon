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
	return evidence.Status, nil
}

func summarize(scenarios []scenario) string {
	for _, scenario := range scenarios {
		if scenario.Status == statusFailed {
			return statusFailed
		}
	}
	return statusPassed
}

func buildScenarios(clientRoot string, cfg config) []scenario {
	out := []scenario{clientReadOnlyScenario(clientRoot)}
	if !*cfg.probeRoutes {
		return out
	}
	out = append(out, loginScenario(*cfg.baseURL))
	out = append(out, workspaceRouteScenarios(*cfg.baseURL, *cfg.workspaceID)...)
	return out
}
