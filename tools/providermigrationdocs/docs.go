package main

func renderedDocs(m manifest) map[string]string {
	docs := map[string]string{m.GeneratedDoc: renderRoot(m)}
	for _, page := range m.Pages {
		docs[page.GeneratedDoc] = renderPage(page)
	}
	return docs
}

func generatedDocPaths(m manifest) []string {
	paths := []string{m.GeneratedDoc}
	for _, page := range m.Pages {
		paths = append(paths, page.GeneratedDoc)
	}
	return paths
}
