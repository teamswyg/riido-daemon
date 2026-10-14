package main

import "strings"

func collectManifestLoopPointerArrayTargets(
	root string,
	ownerPath string,
	object map[string]any,
	fields []string,
	targets map[string]bool,
) {
	for _, field := range fields {
		items, ok := object[field].([]any)
		if !ok {
			continue
		}
		collectManifestLoopPointerArrayItems(root, ownerPath, items, targets)
	}
}

func collectManifestLoopPointerArrayItems(root, ownerPath string, items []any, targets map[string]bool) {
	for _, item := range items {
		source, ok := item.(string)
		if !ok {
			continue
		}
		target, ok := manifestLoopReferencePath(root, ownerPath, source)
		if ok && strings.HasSuffix(target, ".riido.json") {
			targets[target] = true
		}
	}
}
