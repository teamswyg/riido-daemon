package main

import "fmt"

func validateLinks(label string, links []link, want int) []string {
	if len(links) != want {
		return []string{fmt.Sprintf("%s count = %d, want %d", label, len(links), want)}
	}
	var problems []string
	for _, link := range links {
		if link.Title == "" || link.Path == "" {
			problems = append(problems, label+" links require title and path")
		}
	}
	return problems
}
