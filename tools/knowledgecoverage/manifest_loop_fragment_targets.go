package main

import "strings"

func collectManifestLoopFragmentTargets(
	root string,
	ownerPath string,
	fragments map[string]any,
	targets map[string]bool,
) {
	for _, value := range fragments {
		source, ok := value.(string)
		if !ok {
			continue
		}
		target, ok := manifestSiblingSourcePath(root, ownerPath, source)
		if ok && strings.HasSuffix(target, ".riido.json") {
			targets[target] = true
		}
	}
}
