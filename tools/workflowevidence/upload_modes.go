package main

import "strings"

func artifactUploadModes(text string) []string {
	lines := strings.Split(text, "\n")
	var modes []string
	for i, line := range lines {
		if !strings.Contains(line, "actions/upload-artifact") {
			continue
		}
		modes = append(modes, uploadModeFromStep(lines[i+1:]))
	}
	return modes
}

func uploadModeFromStep(lines []string) string {
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") {
			return ""
		}
		if value, ok := strings.CutPrefix(trimmed, "if-no-files-found:"); ok {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func countUploadMode(modes []string, want string) int {
	count := 0
	for _, mode := range modes {
		if mode == want {
			count++
		}
	}
	return count
}

func countNonStrictUploadModes(modes []string) int {
	count := 0
	for _, mode := range modes {
		if mode != "error" {
			count++
		}
	}
	return count
}
