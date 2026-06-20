package main

type detailLink struct {
	Title string
	File  string
}

func detailLinks() []detailLink {
	return []detailLink{
		{Title: "Boundary criteria", File: "boundary-criteria.md"},
		{Title: "Upstream provenance", File: "upstream-provenance.md"},
		{Title: "Screen entries", File: "screen-entries.md"},
		{Title: "Fixture vocabulary", File: "fixture-vocabulary.md"},
		{Title: "Change loop", File: "change-loop.md"},
		{Title: "Verification", File: "verification.md"},
	}
}

func entryScope(m manifest, nodeID string) string {
	for _, entry := range m.Entries {
		if entry.NodeID == nodeID {
			return entry.DaemonScope
		}
	}
	return ""
}
