package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type codexModelCache struct {
	Models []codexModelCacheRecord `json:"models"`
}

type codexModelCacheRecord struct {
	Slug        string `json:"slug"`
	DisplayName string `json:"display_name"`
}

func codexModels() ([]model, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(home, ".codex", "models_cache.json")
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return parseCodexModels(body), nil
}

func parseCodexModels(body []byte) []model {
	var cache codexModelCache
	if err := json.Unmarshal(body, &cache); err != nil {
		return nil
	}
	rows := make([]model, 0, len(cache.Models))
	for _, candidate := range cache.Models {
		if candidate.Slug != "" {
			rows = append(rows, model{ModelID: candidate.Slug, Label: candidate.DisplayName})
		}
	}
	return normalizeModels(rows)
}
