package main

import "fmt"

func generatedDocPaths(m manifest) []string {
	return []string{m.GeneratedDocs.Readme, m.GeneratedDocs.DocumentMap}
}

func validateReadEntry(root string, row readEntry) []string {
	var problems []string
	if row.Doc == "" || row.Description == "" {
		problems = append(problems, "read_order entries require doc and description")
	}
	problems = append(problems, requireDoc(root, row.Doc)...)
	return problems
}

func validateDecision(root string, row decision) []string {
	var problems []string
	if row.Topic == "" || len(row.Docs) == 0 {
		problems = append(problems, "decision entries require topic and docs")
	}
	for _, doc := range row.Docs {
		problems = append(problems, requireDoc(root, doc)...)
	}
	return problems
}

func validateUniqueTopics(rows []decision) []string {
	seen := map[string]bool{}
	var problems []string
	for _, row := range rows {
		if seen[row.Topic] {
			problems = append(problems, fmt.Sprintf("duplicate decision topic %q", row.Topic))
		}
		seen[row.Topic] = true
	}
	return problems
}
