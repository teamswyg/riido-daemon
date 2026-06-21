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
	if ok {
		collectManifestLoopFileMapTargets(root, path, files, targets)
	}
	fragments, ok := object["fragments"].(map[string]any)
	if ok {
		collectManifestLoopFragmentTargets(root, path, fragments, targets)
	}
	entryFiles, ok := object["entry_files"].([]any)
	if ok {
		collectManifestLoopEntryFileTargets(root, path, entryFiles, targets)
	}
}
