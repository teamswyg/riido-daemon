package main

import "strings"

func artifactUploadPathValues(text string) []string {
	lines := strings.Split(text, "\n")
	var paths []string
	for i, line := range lines {
		if strings.Contains(line, "actions/upload-artifact") {
			paths = append(paths, uploadPathValuesFromStep(lines[i+1:])...)
		}
	}
	return uniqueStrings(paths)
}

func uploadPathValuesFromStep(lines []string) []string {
	var paths []string
	for i := range len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "- ") {
			break
		}
		value, ok := strings.CutPrefix(trimmed, "path:")
		if !ok {
			continue
		}
		if strings.TrimSpace(value) == "|" {
			paths = append(paths, uploadPathBlockValues(lines[i+1:], leadingSpaces(lines[i]))...)
			continue
		}
		paths = append(paths, cleanWorkflowValue(value))
	}
	return paths
}

func uploadPathBlockValues(lines []string, pathIndent int) []string {
	var values []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if leadingSpaces(line) <= pathIndent {
			break
		}
		values = append(values, cleanWorkflowValue(line))
	}
	return values
}
