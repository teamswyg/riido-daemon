package main

import (
	"encoding/json"
	"os"
)

func loadJSON[T any](repo, rel string) (T, error) {
	var out T
	path, err := cleanRepoPath(repo, rel)
	if err != nil {
		return out, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return out, err
	}
	return out, json.Unmarshal(data, &out)
}

func readSource(repo, rel string) (string, error) {
	path, err := cleanRepoPath(repo, rel)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	return string(data), err
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
