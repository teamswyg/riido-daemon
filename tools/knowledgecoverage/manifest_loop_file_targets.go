package main

import "strings"

func collectManifestLoopFileMapTargets(
	root string,
	ownerPath string,
	files map[string]any,
	targets map[string]bool,
) {
	for _, value := range files {
		collectManifestLoopFileTargets(root, ownerPath, value, targets)
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
