package main

import (
	"path"
	"slices"
	"strings"
)

func manualTopDirs(docs []docClass, limit int) []manualDir {
	counts := map[string]int{}
	for _, doc := range docs {
		if doc.Kind == "manual_registered" {
			counts[manualDirectory(doc.Path)]++
		}
	}
	out := make([]manualDir, 0, len(counts))
	for dir, count := range counts {
		out = append(out, manualDir{Path: dir, Count: count})
	}
	slices.SortFunc(out, compareManualDir)
	return takeManualDirs(out, limit)
}

func manualDirectory(docPath string) string {
	dir := path.Dir(docPath)
	if strings.HasPrefix(dir, "docs/migration/daemon/") {
		return firstSegments(dir, 4)
	}
	if strings.HasPrefix(dir, "docs/20-domain/") {
		return firstSegments(dir, 3)
	}
	return dir
}

func firstSegments(value string, count int) string {
	parts := strings.Split(value, "/")
	if len(parts) < count {
		return value
	}
	return strings.Join(parts[:count], "/")
}
