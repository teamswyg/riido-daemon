package main

func summarizeLoopRegistry(entries []loopRegistryEntry) loopRegistrySummary {
	out := loopRegistrySummary{
		Count:   len(entries),
		IDs:     make([]string, 0, len(entries)),
		Expires: make([]string, 0, len(entries)),
	}
	for _, entry := range entries {
		out.IDs = append(out.IDs, entry.ID)
		out.Expires = append(out.Expires, entry.ExpiresAfter)
		out.EvidenceRef += len(entry.Evidence)
	}
	return out
}
