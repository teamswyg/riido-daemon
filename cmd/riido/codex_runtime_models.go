package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func codexRuntimeModels(userHome func() (string, error)) []runtimeactor.RuntimeModel {
	modelID := codexConfiguredModelID(userHome)
	if modelID == "" {
		return nil
	}
	return []runtimeactor.RuntimeModel{{ModelID: modelID, Label: modelID, IsDefault: true}}
}

func codexConfiguredModelID(userHome func() (string, error)) string {
	if userHome == nil {
		return ""
	}
	home, err := userHome()
	if err != nil || strings.TrimSpace(home) == "" {
		return ""
	}
	body, err := os.ReadFile(filepath.Join(home, ".codex", "config.toml"))
	if err != nil {
		return ""
	}
	return codexModelValueFromConfig(string(body))
}

func codexModelValueFromConfig(body string) string {
	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, rawValue, ok := strings.Cut(line, "=")
		if ok && strings.TrimSpace(key) == "model" {
			return parseCodexModelValue(rawValue)
		}
	}
	return ""
}

func parseCodexModelValue(rawValue string) string {
	value := strings.TrimSpace(rawValue)
	if unquoted, err := strconv.Unquote(value); err == nil {
		return strings.TrimSpace(unquoted)
	}
	if commentAt := strings.Index(value, "#"); commentAt >= 0 {
		value = strings.TrimSpace(value[:commentAt])
		if unquoted, err := strconv.Unquote(value); err == nil {
			return strings.TrimSpace(unquoted)
		}
	}
	return strings.TrimSpace(value)
}
