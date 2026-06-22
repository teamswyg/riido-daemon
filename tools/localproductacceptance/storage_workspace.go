package main

import "strings"

const workspaceKeySuffix = "-toggle"

func stateWorkspaceID(state storageState) string {
	for _, origin := range state.Origins {
		for _, entry := range origin.LocalStorage {
			if id, ok := parseWorkspaceKey(entry.Name); ok {
				return id
			}
		}
	}
	return ""
}

func parseWorkspaceKey(key string) (string, bool) {
	if !strings.HasPrefix(key, "workspace-") {
		return "", false
	}
	if !strings.HasSuffix(key, workspaceKeySuffix) {
		return "", false
	}
	id := strings.TrimSuffix(strings.TrimPrefix(key, "workspace-"), workspaceKeySuffix)
	return id, id != ""
}
