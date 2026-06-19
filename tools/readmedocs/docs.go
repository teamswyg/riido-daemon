package main

func generatedDocPaths(m manifest) []string {
	var paths []string
	for _, page := range m.Pages {
		paths = append(paths, page.GeneratedDoc)
	}
	return paths
}
