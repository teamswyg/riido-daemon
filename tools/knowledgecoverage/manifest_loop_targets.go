package main

import (
	"os"
	"path/filepath"
	"strings"
)

func scanManifestLoopDelegatedTargets(root string) (map[string]bool, error) {
	targets := map[string]bool{}
	for {
		changed, err := scanManifestLoopDelegatedTargetPass(root, targets)
		if err != nil || !changed {
			return targets, err
		}
	}
}

func scanManifestLoopDelegatedTargetPass(root string, targets map[string]bool) (bool, error) {
	changed := false
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
		if manifestLoopStatus(root, path) == "missing" && !targets[path] {
			return nil
		}
		before := len(targets)
		collectManifestLoopDelegatedTargets(root, path, targets)
		changed = changed || len(targets) > before
		return nil
	})
	return changed, err
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
	collectManifestLoopPointerTargets(root, path, object, manifestLoopPointerFields(), targets)
	collectManifestLoopPointerArrayTargets(root, path, object, manifestLoopPointerArrayFields(), targets)
}
