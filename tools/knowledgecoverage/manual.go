package main

import "strings"

func manualPathIndex(m manifest) map[string]manualGroup {
	result := map[string]manualGroup{}
	for _, group := range m.ManualGroups {
		for _, path := range group.Paths {
			result[path] = group
		}
	}
	return result
}

func manualPrefixMatch(m manifest, path string) (manualGroup, bool) {
	for _, group := range m.ManualGroups {
		for _, prefix := range group.PathPrefixes {
			if strings.HasPrefix(path, prefix) {
				return group, true
			}
		}
	}
	return manualGroup{}, false
}

func manualGroupIDs(m manifest) []string {
	ids := make([]string, 0, len(m.ManualGroups))
	for _, group := range m.ManualGroups {
		ids = append(ids, group.ID)
	}
	return ids
}
