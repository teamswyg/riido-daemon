package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

type openClawConfig struct {
	Agents openClawConfigAgents `json:"agents"`
	Models openClawConfigModels `json:"models"`
}

type openClawConfigAgents struct {
	Defaults openClawConfigDefaults `json:"defaults"`
}

type openClawConfigDefaults struct {
	Model openClawConfigDefaultModel `json:"model"`
}

type openClawConfigDefaultModel struct {
	Primary string `json:"primary"`
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

func openClawRuntimeModels(userHome func() (string, error)) []runtimeactor.RuntimeModel {
	if userHome == nil {
		return nil
	}
	home, err := userHome()
	if err != nil || strings.TrimSpace(home) == "" {
		return nil
	}
	body, err := os.ReadFile(filepath.Join(home, ".openclaw", "openclaw.json"))
	if err != nil {
		return nil
	}
	return parseOpenClawRuntimeModels(body)
}

func parseOpenClawRuntimeModels(body []byte) []runtimeactor.RuntimeModel {
	var cfg openClawConfig
	if err := json.Unmarshal(body, &cfg); err != nil {
		return nil
	}
	defaultID := strings.TrimSpace(cfg.Agents.Defaults.Model.Primary)
	models := openClawConfiguredModels(cfg)
	if len(models) == 0 {
		model, ok := runtimeModelRecord(defaultID, defaultID, true)
		if ok {
			return []runtimeactor.RuntimeModel{model}
		}
	}
	return normalizeRuntimeModels(models, defaultID)
}
