package main

import (
	"fmt"
	"os"
	"strings"
)

func validateRequiredNoticeTerms(repoRoot string, terms []string) []string {
	if len(terms) == 0 {
		return nil
	}
	data, err := os.ReadFile(resolvePath(repoRoot, "NOTICE.md"))
	if err != nil {
		return []string{fmt.Sprintf("read NOTICE.md: %v", err)}
	}
	return missingNoticeTerms(string(data), terms)
}

func missingNoticeTerms(text string, terms []string) []string {
	var problems []string
	for _, term := range terms {
		trimmed := strings.TrimSpace(term)
		if trimmed == "" {
			continue
		}
		if !strings.Contains(text, trimmed) {
			problems = append(problems, fmt.Sprintf("NOTICE.md must include %q", trimmed))
		}
	}
	return problems
}
