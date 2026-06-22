package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type figmaIntentEntry struct {
	NodeID           string   `json:"node_id"`
	Name             string   `json:"name"`
	UpstreamOwner    []string `json:"upstream_owner"`
	DaemonScope      string   `json:"daemon_scope"`
	DaemonConsumed   []string `json:"daemon_consumed_facts"`
	ClientOwnedFacts []string `json:"client_owned_facts"`
}

func figmaIntentScenarios(path, goldenPath, screenshotDir string) []scenario {
	entries, err := loadFigmaIntentEntries(path)
	if err != nil {
		return failedFigmaIntentScenarios(path, err)
	}
	goldens, goldenErr := loadFigmaGoldenCatalog(goldenPath)
	return []scenario{
		figmaCatalogScenario(path, entries),
		figmaScreenScenario("figma.onboarding", entries, "온보딩", goldens, goldenErr, screenshotDir),
		figmaScreenScenario("figma.runtime.settings", entries, "런타임 설정페이지", goldens, goldenErr, screenshotDir),
	}
}

func loadFigmaIntentEntries(path string) ([]figmaIntentEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read figma intent manifest: %w", err)
	}
	var entries []figmaIntentEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("decode figma intent manifest: %w", err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("figma intent manifest has no entries")
	}
	return entries, nil
}
