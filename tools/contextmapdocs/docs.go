package main

func renderedDocs(m manifest) map[string]string {
	return map[string]string{
		m.GeneratedDoc: renderRoot(m),
		"docs/20-domain/context-map/bounded-contexts.md":            renderBoundedContexts(m),
		"docs/20-domain/context-map/dependency-direction.md":        renderDependency(m),
		"docs/20-domain/context-map/acl-locations.md":               renderACL(m),
		"docs/20-domain/context-map/split-repo-ownership.md":        renderSplitRepo(m),
		"docs/20-domain/context-map/figma-daemon-boundaries.md":     renderFigmaDaemon(m),
		"docs/20-domain/context-map/figma-onboarding-boundaries.md": renderFigmaOnboarding(m),
		"docs/20-domain/context-map/change-procedure.md":            renderChangeProcedure(m),
	}
}

func generatedDocPaths(m manifest) []string {
	return []string{
		m.GeneratedDoc,
		"docs/20-domain/context-map/bounded-contexts.md",
		"docs/20-domain/context-map/dependency-direction.md",
		"docs/20-domain/context-map/acl-locations.md",
		"docs/20-domain/context-map/split-repo-ownership.md",
		"docs/20-domain/context-map/figma-daemon-boundaries.md",
		"docs/20-domain/context-map/figma-onboarding-boundaries.md",
		"docs/20-domain/context-map/change-procedure.md",
	}
}
