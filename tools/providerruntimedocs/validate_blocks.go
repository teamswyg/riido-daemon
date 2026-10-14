package main

func validateItems(id string, items []string) []string {
	if len(items) == 0 {
		return []string{id + " has empty list"}
	}
	return nil
}

func validateLinks(id string, links []link) []string {
	if len(links) == 0 {
		return []string{id + " has empty links"}
	}
	var problems []string
	for _, link := range links {
		if link.Title == "" || link.Path == "" {
			problems = append(problems, id+" links require title and path")
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
