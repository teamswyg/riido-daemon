package main

import "strings"

func collectManifestLoopEntryFileTargets(root, ownerPath string, files []any, targets map[string]bool) {
	for _, item := range files {
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
