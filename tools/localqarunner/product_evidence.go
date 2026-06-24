package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type productRunEvidence struct {
	Scenarios []productRunScenario `json:"scenarios"`
}

type productRunScenario struct {
	Observed productRunObserved `json:"observed"`
}

type productRunObserved struct {
	ClosedLoops []runLoopCandidate `json:"closed_loop_candidates"`
}

func applyProductEvidence(root string, cfg config, evidence *runEvidence) error {
	path := outputPath(root, *cfg.productEvidence)
	if !fileExists(path) {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read product evidence: %w", err)
	}
	var file productRunEvidence
	if err := json.Unmarshal(data, &file); err != nil {
		return fmt.Errorf("parse product evidence: %w", err)
	}
	if err := validateProductClosedLoops(file.Scenarios); err != nil {
		return err
	}
	appendClosedLoops(evidence, file.Scenarios)
	return nil
}

func validateProductClosedLoops(scenarios []productRunScenario) error {
	for _, scenario := range scenarios {
		for _, candidate := range scenario.Observed.ClosedLoops {
			if len(candidate.RequiredNextArtifacts) == 0 {
				return fmt.Errorf("closed-loop candidate %s missing required_next_artifacts", candidate.ID)
			}
		}
	}
	return nil
}

func appendClosedLoops(evidence *runEvidence, scenarios []productRunScenario) {
	for _, scenario := range scenarios {
		evidence.ClosedLoops = append(evidence.ClosedLoops, scenario.Observed.ClosedLoops...)
	}
	if len(evidence.ClosedLoops) > 0 {
		evidence.CoverageStatus = mergeCoverageStatus(evidence.CoverageStatus, statusPartial)
	}
}
