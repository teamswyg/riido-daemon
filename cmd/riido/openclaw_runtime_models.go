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
	modelID := strings.TrimSpace(cfg.Agents.Defaults.Model.Primary)
	model, ok := runtimeModelRecord(modelID, modelID, true)
	if !ok {
		return nil
	}
	return []runtimeactor.RuntimeModel{model}
}
