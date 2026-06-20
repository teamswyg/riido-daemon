package main

func renderedDocs(m manifest) map[string]string {
	return map[string]string{
		"docs/30-architecture/figma-ai-agent-daemon-boundary.md":                     renderRoot(m),
		"docs/30-architecture/figma-ai-agent-daemon-boundary/boundary-criteria.md":   renderBoundaryCriteria(m),
		"docs/30-architecture/figma-ai-agent-daemon-boundary/change-loop.md":         renderChangeLoop(m),
		"docs/30-architecture/figma-ai-agent-daemon-boundary/fixture-vocabulary.md":  renderFixtureVocabulary(m),
		"docs/30-architecture/figma-ai-agent-daemon-boundary/screen-entries.md":      renderScreenEntries(m),
		"docs/30-architecture/figma-ai-agent-daemon-boundary/upstream-provenance.md": renderUpstreamProvenance(m),
		"docs/30-architecture/figma-ai-agent-daemon-boundary/verification.md":        renderVerification(m),
	}
}

func generatedDocPaths() []string {
	return []string{
		"docs/30-architecture/figma-ai-agent-daemon-boundary.md",
		"docs/30-architecture/figma-ai-agent-daemon-boundary/boundary-criteria.md",
		"docs/30-architecture/figma-ai-agent-daemon-boundary/change-loop.md",
		"docs/30-architecture/figma-ai-agent-daemon-boundary/fixture-vocabulary.md",
		"docs/30-architecture/figma-ai-agent-daemon-boundary/screen-entries.md",
		"docs/30-architecture/figma-ai-agent-daemon-boundary/upstream-provenance.md",
		"docs/30-architecture/figma-ai-agent-daemon-boundary/verification.md",
	}
}
