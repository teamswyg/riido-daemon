package main

func validateBlocks(detail detailDoc) []string {
	var problems []string
	for _, block := range detail.Blocks {
		switch block.Kind {
		case "paragraph":
			if block.Text == "" {
				problems = append(problems, detail.ID+" has empty paragraph")
			}
		case "bullets", "ordered":
			if len(block.Items) == 0 {
				problems = append(problems, detail.ID+" has empty list")
			}
		case "table":
			problems = append(problems, validateTable(detail.ID, block)...)
		case "code":
			if block.Code == "" {
				problems = append(problems, detail.ID+" has empty code block")
			}
		default:
			problems = append(problems, detail.ID+" has invalid block kind "+block.Kind)
		}
	}
	return problems
}

func validateTable(id string, block block) []string {
	if len(block.Columns) == 0 || len(block.Rows) == 0 {
		return []string{id + " has empty table"}
	}
	var problems []string
	for _, row := range block.Rows {
		if len(row) != len(block.Columns) {
			problems = append(problems, id+" table row width mismatch")
		}
	}
	return problems
}
