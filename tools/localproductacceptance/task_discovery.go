package main

import "sort"

func firstAssignedProfileTaskID(payload map[string]any) string {
	profiles, _ := payload["assigned_agent_profiles"].(map[string]any)
	if len(profiles) == 0 {
		return ""
	}
	keys := make([]string, 0, len(profiles))
	for key := range profiles {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys[0]
}
