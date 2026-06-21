package main

import "strings"

func manifestLoopPointerFields() []string {
	return []string{
		"import_rules_file",
		"package_roles_file",
		"ports_file",
		"provider_validation_manifest",
		"real_cli_observation_manifest",
	}
}

func collectManifestLoopPointerTargets(
	root string,
	ownerPath string,
	object map[string]any,
	fields []string,
	targets map[string]bool,
) {
	for _, field := range fields {
		source, ok := object[field].(string)
		if !ok {
			continue
		}
		target, ok := manifestSiblingSourcePath(root, ownerPath, source)
		if ok && strings.HasSuffix(target, ".riido.json") {
			targets[target] = true
		}
	}
}
