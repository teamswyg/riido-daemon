package main

func renderedDocs(m manifest) map[string]string {
	return map[string]string{
		m.GeneratedDoc:       renderRoot(m),
		m.DetailDocs[0].Path: renderAssets(m),
		m.DetailDocs[1].Path: renderInstaller(m),
		m.DetailDocs[2].Path: renderDesktopMSIX(m),
		m.DetailDocs[3].Path: renderReviewBoundary(m),
	}
}

func generatedDocPaths(m manifest) []string {
	return []string{
		m.GeneratedDoc,
		m.DetailDocs[0].Path,
		m.DetailDocs[1].Path,
		m.DetailDocs[2].Path,
		m.DetailDocs[3].Path,
	}
}
