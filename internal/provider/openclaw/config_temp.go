package openclaw

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func maybeWriteTaskScopedConfig(req agentbridge.StartRequest, env map[string]string) ([]string, error) {
	if strings.TrimSpace(req.Cwd) == "" {
		return nil, nil
	}
	source, ok := sourceConfigPath(env)
	if !ok {
		return nil, nil
	}
	raw, ok, err := readOptionalConfig(source)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	var config map[string]any
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, fmt.Errorf("openclaw: parse config for task workspace overlay: %w", err)
	}
	applyTaskScopedConfig(config, req.Cwd, req.Model)
	path, err := writeTempConfig(config)
	if err != nil {
		return nil, err
	}
	env[openClawConfigPathEnv] = path
	return []string{path}, nil
}

func readOptionalConfig(path string) ([]byte, bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("openclaw: read config for task workspace overlay: %w", err)
	}
	return raw, true, nil
}

func writeTempConfig(config map[string]any) (string, error) {
	file, err := os.CreateTemp("", "riido-openclaw-*.json")
	if err != nil {
		return "", fmt.Errorf("openclaw: create temp config: %w", err)
	}
	path := file.Name()
	if err := json.NewEncoder(file).Encode(config); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("openclaw: write temp config: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("openclaw: close temp config: %w", err)
	}
	return path, nil
}
