package main

import (
	"os"
	"path/filepath"
	"strings"
)

func scanManifestLoopDelegatedTargets(root string) (map[string]bool, error) {
	targets := map[string]bool{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && filepath.Base(path) == ".git" {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".riido.json") {
			return nil
		}
		if manifestLoopStatus(root, path) == "missing" {
			return nil
		}
		collectManifestLoopDelegatedTargets(root, path, targets)
		return nil
	})
	return targets, err
}

func collectManifestLoopDelegatedTargets(root, path string, targets map[string]bool) {
	object, ok := readManifestObject(path)
	if !ok {
		return
	}
	files, ok := object["evidence_files"].(map[string]any)
	if !ok {
		return
	}
	for _, value := range files {
		collectManifestLoopFileTargets(root, path, value, targets)
	}
}

func collectManifestLoopFileTargets(root, ownerPath string, value any, targets map[string]bool) {
	items, ok := value.([]any)
	if !ok {
		return
	}
	for _, item := range items {
		source, ok := item.(string)
		if !ok {
			continue
		}
		target, ok := manifestSiblingSourcePath(root, ownerPath, source)
		if ok && strings.HasSuffix(target, ".riido.json") {
			targets[target] = true
		}
	}
}
