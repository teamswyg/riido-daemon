package main

func checkDetailDocs(manifest Manifest) []problem {
	var problems []problem
	seen := map[string]bool{manifest.GeneratedDoc: true}
	for _, doc := range manifest.DetailDocs {
		problems = append(problems, validateDetailDoc(manifest, doc, seen)...)
	}
	return problems
}

func validateDetailDoc(manifest Manifest, doc DetailDoc, seen map[string]bool) []problem {
	var problems []problem
	if doc.Title == "" || doc.Path == "" || len(doc.Blocks) == 0 {
		problems = append(problems, problem{Message: "detail doc title, path, and blocks are required"})
	}
	if seen[doc.Path] {
		problems = append(problems, problem{Message: "duplicate generated doc path: " + doc.Path})
	}
	seen[doc.Path] = true
	for _, block := range doc.Blocks {
		problems = append(problems, validateDetailBlock(manifest, doc.Path, block)...)
	}
	return problems
}
