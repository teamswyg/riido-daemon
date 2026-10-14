package main

func checkDetailDocs(manifest Manifest) []problem {
	var problems []problem
	seen := map[string]bool{manifest.GeneratedDoc: true}
	for _, doc := range manifest.DetailDocs {
		problems = append(problems, validateDetailDoc(doc, seen)...)
	}
	return problems
}

func validateDetailDoc(doc DetailDoc, seen map[string]bool) []problem {
	var problems []problem
	if doc.Title == "" || doc.Path == "" || len(doc.Blocks) == 0 {
		problems = append(problems, problem{Message: "detail doc title, path, and blocks are required"})
	}
	if seen[doc.Path] {
		problems = append(problems, problem{Message: "duplicate generated doc path: " + doc.Path})
	}
	seen[doc.Path] = true
	for _, block := range doc.Blocks {
		if !validDetailBlock(block) {
			problems = append(problems, problem{Message: "invalid detail doc block in " + doc.Path})
		}
	}
	return problems
}

func validDetailBlock(block DetailBlock) bool {
	switch block.Kind {
	case "paragraph":
		return block.Text != ""
	case "bullets", "commands":
		return len(block.Items) > 0
	case "command_groups_table":
		return true
	default:
		return false
	}
}
