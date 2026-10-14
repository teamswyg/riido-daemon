package main

func figmaObserved(entries []figmaIntentEntry) map[string]any {
	return map[string]any{
		"matches_count": len(entries),
		"entries":       figmaObservedEntries(entries),
	}
}

func figmaObservedEntries(entries []figmaIntentEntry) []map[string]any {
	out := make([]map[string]any, 0, len(entries))
	for _, entry := range entries {
		out = append(out, map[string]any{
			"node_id":               entry.NodeID,
			"name":                  entry.Name,
			"upstream_owner":        entry.UpstreamOwner,
			"daemon_consumed_facts": entry.DaemonConsumed,
			"client_owned_facts":    entry.ClientOwnedFacts,
			"daemon_scope":          entry.DaemonScope,
		})
	}
	return out
}

func figmaScreenNames(entries []figmaIntentEntry) []string {
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		out = append(out, entry.Name)
	}
	return out
}
