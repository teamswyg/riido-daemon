package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func loadFigmaGoldenCatalog(path string) (map[string]figmaGoldenScreen, error) {
	if path == "" {
		return nil, fmt.Errorf("figma golden manifest path is empty")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read figma golden manifest: %w", err)
	}
	var catalog figmaGoldenCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("decode figma golden manifest: %w", err)
	}
	if len(catalog.Screens) == 0 {
		return nil, fmt.Errorf("figma golden manifest has no screens")
	}
	return indexFigmaGoldens(filepath.Dir(path), catalog.Screens), nil
}

func indexFigmaGoldens(base string, screens []figmaGoldenScreen) map[string]figmaGoldenScreen {
	out := make(map[string]figmaGoldenScreen, len(screens))
	for _, screen := range screens {
		screen.ResolvedPath = filepath.Join(base, screen.GoldenPath)
		out[screen.ScenarioID] = screen
	}
	return out
}
