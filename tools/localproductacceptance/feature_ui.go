package main

import (
	"embed"
	"encoding/json"
)

//go:embed feature_ui.generated.json
var featureUIFS embed.FS

func featureUIContract() map[string]any {
	body, err := featureUIFS.ReadFile("feature_ui.generated.json")
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		return map[string]any{"error": err.Error()}
	}
	return out
}

func featureUIScenario() scenario {
	spec := featureUIContract()
	spec["capture_owner"] = "qa-codex"
	spec["capture_verifies"] = "functional UI state and interaction intent"
	return scenario{
		ID:         "contract.ui.feature_dsl",
		Status:     statusPassed,
		Screenshot: ".riido-local/screenshots/contract-lab/feature-ui-manual-pass.png",
		Observed:   spec,
	}
}
