package main

func validatePage(p page, seen map[string]bool) []string {
	var problems []string
	if p.SchemaVersion != pageSchema {
		problems = append(problems, "unexpected page schema_version: "+p.ID)
	}
	if p.ID == "" || p.Title == "" || p.GeneratedDoc == "" || len(p.Blocks) == 0 {
		problems = append(problems, "page id, title, generated_doc, and blocks are required")
	}
	if seen[p.ID] {
		problems = append(problems, "duplicate page id "+p.ID)
	}
	seen[p.ID] = true
	for _, block := range p.Blocks {
		problems = append(problems, validateBlock(p.ID, block)...)
	}
	return problems
}

func validateBlock(id string, block block) []string {
	switch block.Kind {
	case "heading", "paragraph":
		if block.Text == "" {
			return []string{id + " has empty text block"}
		}
	case "bullets":
		return validateItems(id, block.Items)
	case "links":
		return validateLinks(id, block.Links)
	case "table":
		return validateTable(id, block)
	case "code":
		if block.Code == "" || block.Language == "" {
			return []string{id + " has empty code block"}
		}
	default:
		return []string{id + " has invalid block kind " + block.Kind}
	}
	return nil
}
