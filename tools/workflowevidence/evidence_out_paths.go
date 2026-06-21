package main

import (
	"regexp"
	"strings"
)

var evidenceOutPattern = regexp.MustCompile(`-evidence-out(?:=|\s+)("[^"]+"|'[^']+'|[^\s\\]+)`)

func evidenceOutPaths(text string) []string {
	matches := evidenceOutPattern.FindAllStringSubmatch(text, -1)
	paths := make([]string, 0, len(matches))
	for _, match := range matches {
		paths = append(paths, cleanWorkflowValue(match[1]))
	}
	return uniqueStrings(paths)
}

func cleanWorkflowValue(value string) string {
	return strings.Trim(strings.TrimSpace(value), `"'`)
}
