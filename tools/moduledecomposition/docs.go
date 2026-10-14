package main

func renderedDocs(m manifest) map[string]string {
	return map[string]string{
		m.GeneratedDoc:       renderRoot(m),
		m.DetailDocs[0].Path: renderPackageMap(m),
		m.DetailDocs[1].Path: renderImportRules(m),
		m.DetailDocs[2].Path: renderHexagonalPorts(m),
		m.DetailDocs[3].Path: renderFactorBoundary(m),
		m.DetailDocs[4].Path: renderChangeProcedure(m),
	}
}

func generatedDocPaths(m manifest) []string {
	return []string{
		m.GeneratedDoc,
		m.DetailDocs[0].Path,
		m.DetailDocs[1].Path,
		m.DetailDocs[2].Path,
		m.DetailDocs[3].Path,
		m.DetailDocs[4].Path,
	}
}
