package main

import (
	"strings"
)

func generatedToolPath(generator string) (string, bool) {
	for field := range strings.FieldsSeq(generator) {
		if strings.HasPrefix(field, "./tools/") {
			return strings.TrimPrefix(field, "./"), true
		}
		if strings.HasPrefix(field, "tools/") {
			return field, true
		}
	}
	return "", false
}

func isWorkflowFile(path string) bool {
	return strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml")
}
