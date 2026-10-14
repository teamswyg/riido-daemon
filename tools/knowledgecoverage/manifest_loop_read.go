package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

func readManifestObject(path string) (map[string]any, bool) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var object map[string]any
	if err := json.Unmarshal(body, &object); err != nil {
		return nil, false
	}
	return object, true
}

func manifestDocHasLoop(path string) bool {
	object, ok := readManifestObject(path)
	if !ok {
		return false
	}
	loop, ok := object["loop"].(map[string]any)
	if !ok {
		return false
	}
	for _, key := range []string{"observation", "hypothesis", "execute", "evaluate", "retrospective"} {
		value, ok := loop[key].(string)
		if !ok || strings.TrimSpace(value) == "" {
			return false
		}
	}
	return true
}

func manifestLoopSourcePath(root, path string) (string, bool) {
	object, ok := readManifestObject(path)
	if !ok {
		return "", false
	}
	source, ok := object["loop_source"].(string)
	if !ok || strings.TrimSpace(source) == "" {
		return "", false
	}
	return manifestSourcePath(root, source)
}

func manifestSourcePath(root, source string) (string, bool) {
	path := filepath.Join(root, filepath.FromSlash(source))
	if filepath.IsAbs(source) {
		path = source
	}
	rel, err := filepath.Rel(root, path)
	if err != nil || filepath.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return filepath.Join(root, rel), true
}
