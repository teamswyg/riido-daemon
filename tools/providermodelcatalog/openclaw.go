package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type openClawConfig struct {
	Models openClawConfigModels `json:"models"`
}

type openClawConfigModels struct {
	Providers map[string]openClawConfigProvider `json:"providers"`
}

type openClawConfigProvider struct {
	Models []openClawConfigModel `json:"models"`
}

type openClawConfigModel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func openClawModels() ([]model, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(home, ".openclaw", "openclaw.json")
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return parseOpenClawModels(body), nil
}

func parseOpenClawModels(body []byte) []model {
	var cfg openClawConfig
	if err := json.Unmarshal(body, &cfg); err != nil {
		return nil
	}
	rows := make([]model, 0)
	for provider, entry := range cfg.Models.Providers {
		rows = append(rows, openClawProviderModels(provider, entry)...)
	}
	return normalizeModels(rows)
}

func openClawProviderModels(provider string, entry openClawConfigProvider) []model {
	rows := make([]model, 0, len(entry.Models))
	for _, candidate := range entry.Models {
		rows = append(rows, model{
			ModelID: openClawQualifiedModelID(provider, candidate.ID),
			Label:   candidate.Name,
		})
	}
	return rows
}

func openClawQualifiedModelID(provider, modelID string) string {
	provider = strings.TrimSpace(provider)
	modelID = strings.TrimSpace(modelID)
	if provider == "" || strings.Contains(modelID, "/") {
		return modelID
	}
	return provider + "/" + modelID
}
