package main

import (
	"fmt"
	"os"
	"strings"
)

func validateManualGroupPrefixes(group manualGroup, seen map[string]string) []string {
	var problems []string
	for _, prefix := range group.PathPrefixes {
		if prefix == "" {
			problems = append(problems, fmt.Sprintf("manual group %q has empty path_prefix", group.ID))
			continue
		}
		if existing, ok := seen[prefix]; ok {
			problems = append(problems, fmt.Sprintf("manual prefix %q appears in %q and %q", prefix, existing, group.ID))
		}
		seen[prefix] = group.ID
	}
	return problems
}

func validateManualPrefix(root, groupID, prefix string, docs []docClass) []string {
	if _, err := os.Stat(resolvePath(root, strings.TrimSuffix(prefix, "/"))); err != nil {
		return []string{fmt.Sprintf("manual group %q prefix %q missing", groupID, prefix)}
	}
	for _, doc := range docs {
		if strings.HasPrefix(doc.Path, prefix) && doc.Kind == "manual_registered" && doc.Group == groupID {
			return nil
		}
	}
	return []string{fmt.Sprintf("manual group %q prefix %q matched no manual docs", groupID, prefix)}
}
