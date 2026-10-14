package main

type renderedDocs struct {
	Allowed   string
	Forbidden string
}

type renderedDoc struct {
	path string
	body string
}

func renderAll(manifest Manifest) renderedDocs {
	return renderedDocs{
		Allowed:   renderAllowed(manifest),
		Forbidden: renderForbidden(manifest),
	}
}

func docPairs(manifest Manifest, docs renderedDocs) []renderedDoc {
	return []renderedDoc{
		{path: manifest.AllowedDoc, body: docs.Allowed},
		{path: manifest.ForbiddenDoc, body: docs.Forbidden},
	}
}
