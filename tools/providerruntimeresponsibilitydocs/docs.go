package main

func renderedDocs(m manifest) map[string]string {
	docs := map[string]string{m.GeneratedDoc: renderRoot(m)}
	for _, detail := range m.Details {
		docs[detail.GeneratedDoc] = renderDetail(detail)
	}
	return docs
}

func generatedDocPaths(m manifest) []string {
	paths := []string{m.GeneratedDoc}
	for _, detail := range m.Details {
		paths = append(paths, detail.GeneratedDoc)
	}
	return paths
}
