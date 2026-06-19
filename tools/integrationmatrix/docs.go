package main

func renderedDocs(m manifest) map[string]string {
	return map[string]string{
		m.GeneratedDoc:       renderRoot(m),
		m.DetailDocs[0].Path: renderGatePolicy(m),
		m.DetailDocs[1].Path: renderProviderMatrix(m),
		m.DetailDocs[2].Path: renderAssertions(m),
		m.DetailDocs[3].Path: renderInstructionEffectiveness(m),
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
