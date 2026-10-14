package main

func generatedDocPaths(m manifest) []string {
	paths := []string{m.GeneratedDoc}
	for _, page := range m.Pages {
		paths = append(paths, page.GeneratedDoc)
	}
	return paths
}
