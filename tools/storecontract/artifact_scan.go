package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func scanArtifactRoots(repoRoot string, roots, providerNames []string) []string {
	var problems []string
	for _, root := range roots {
		path := resolvePath(repoRoot, root)
		info, err := os.Stat(path)
		if err != nil {
			problems = append(problems, fmt.Sprintf("store artifact root missing: %s", root))
			continue
		}
		if !info.IsDir() {
			problems = append(problems, fmt.Sprintf("store artifact root is not a directory: %s", root))
			continue
		}
		walkErr := filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				problems = append(problems, fmt.Sprintf("scan %s: %v", path, err))
				return nil
			}
			if entry.IsDir() {
				return nil
			}
			if matchesProviderBinary(entry.Name(), providerNames) {
				problems = append(problems, fmt.Sprintf("provider CLI appears bundled in store artifact root: %s", path))
			}
			if hasHardcodedUserPath(path) {
				problems = append(problems, fmt.Sprintf("store artifact contains hardcoded user path: %s", path))
			}
			return nil
		})
		if walkErr != nil {
			problems = append(problems, fmt.Sprintf("scan root %s: %v", root, walkErr))
		}
	}
	return problems
}

func matchesProviderBinary(filename string, providerNames []string) bool {
	base := strings.ToLower(filename)
	ext := strings.ToLower(filepath.Ext(base))
	stem := strings.TrimSuffix(base, ext)
	executableExt := ext == "" || ext == ".exe" || ext == ".cmd" || ext == ".bat" || ext == ".ps1" || ext == ".sh"
	if !executableExt {
		return false
	}
	for _, provider := range providerNames {
		name := strings.ToLower(provider)
		if base == name || stem == name {
			return true
		}
	}
	return false
}

func hasHardcodedUserPath(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	text := string(data)
	for _, marker := range []string{"/Users/", `C:\Users\`, "C:/Users/", "~/Library/LaunchAgents", "~/Library/Application Support"} {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

func contains(items []string, wanted string) bool {
	return slices.Contains(items, wanted)
}

func resolvePath(repoRoot, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(repoRoot, path)
}

func writeJSON(path string, value checkResult) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode check output: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write check output: %w", err)
	}
	return nil
}
