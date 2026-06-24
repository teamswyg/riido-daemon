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
		Scenarios:     buildScenarios(clientRoot, cfg, observed),
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
		if scenario.Status == statusSkipped || scenario.Status == statusPartial {
			status = statusPartial
		}
	}
	return status
}

func buildScenarios(clientRoot string, cfg config, observed time.Time) []scenario {
	out := []scenario{clientReadOnlyScenario(clientRoot)}
	contract := contractAPIScenarios(cfg)
	out = append(out, contract...)
	domainRows := domainJourneyScenarios(cfg, contract)
	out = append(out, domainRows...)
	out = append(out, stagingFixtureHandoffScenario(cfg, domainRows))
	out = append(out, figmaIntentScenarios(*cfg.figmaManifest, *cfg.figmaGolden, *cfg.screenshots)...)
	out = append(out, figmaRefreshScenario(observed, *cfg.validFor, *cfg.figmaManifest, *cfg.figmaGolden))
	out = append(out, qaLoopScenario(*cfg.validFor, *cfg.figmaManifest, *cfg.labOut, *cfg.manualOut))
	out = append(out, featureUIScenario())
	out = append(out, qaI18NScenario())
	out = append(out, browserMeaningScenario())
	out = append(out, qaSystemScenario())
	out = append(out, evidenceGapScenario(out, cfg))
	out = append(out, contractUIScenario(*cfg.labOut, *cfg.manualOut))
	return out
}
