package main

import (
	"fmt"
	"os"
)

func validateManualGroupPaths(group manualGroup, seen map[string]string) []string {
	var problems []string
	for _, path := range group.Paths {
		if path == "" {
			problems = append(problems, fmt.Sprintf("manual group %q has empty path", group.ID))
			continue
		}
		if existing, ok := seen[path]; ok {
			problems = append(problems, fmt.Sprintf("manual path %q appears in %q and %q", path, existing, group.ID))
		}
		seen[path] = group.ID
	}
	return problems
}

func validateManualEntries(root string, m manifest, docs []docClass) []string {
	var problems []string
	classes := map[string]docClass{}
	for _, doc := range docs {
		classes[doc.Path] = doc
		if doc.Kind == "unregistered_manual" {
			problems = append(problems, fmt.Sprintf("unregistered manual architecture doc %q", doc.Path))
		}
	}
	for _, group := range m.ManualGroups {
		for _, path := range group.Paths {
			problems = append(problems, validateManualPath(root, group.ID, path, classes)...)
		}
		for _, prefix := range group.PathPrefixes {
			problems = append(problems, validateManualPrefix(root, group.ID, prefix, docs)...)
		}
	}
	return problems
}

func validateManualPath(root, groupID, path string, classes map[string]docClass) []string {
	if _, err := os.Stat(resolvePath(root, path)); err != nil {
		return []string{fmt.Sprintf("manual group %q path %q missing", groupID, path)}
	}
	if class, ok := classes[path]; ok && class.Kind != "manual_registered" {
		return []string{fmt.Sprintf("manual group %q path %q is %s", groupID, path, class.Kind)}
	}
	return nil
}
