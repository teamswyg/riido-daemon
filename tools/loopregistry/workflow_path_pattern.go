package main

import "strings"

func workflowPathPatterns(body string) []string {
	var out []string
	for line := range strings.SplitSeq(body, "\n") {
		pattern, ok := workflowPathPattern(line)
		if ok {
			out = append(out, slash(pattern))
		}
	}
	return out
}

func workflowPathPattern(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "- ") {
		return "", false
	}
	value := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
	value = strings.Trim(value, `"'`)
	if value == "" || strings.Contains(value, "${{") {
		return "", false
	}
	return value, true
}

func workflowCoversPath(patterns []string, rel string) bool {
	rel = slash(rel)
	for _, pattern := range patterns {
		if pathPatternCovers(pattern, rel) {
			return true
		}
	}
	return false
}

func pathPatternCovers(pattern, rel string) bool {
	if pattern == rel {
		return true
	}
	if prefix, ok := strings.CutSuffix(pattern, "/**"); ok {
		return rel == prefix || strings.HasPrefix(rel, prefix+"/")
	}
	if prefix, ok := strings.CutSuffix(pattern, "*"); ok {
		return strings.HasPrefix(rel, prefix)
	}
	return false
}
