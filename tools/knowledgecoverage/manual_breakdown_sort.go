package main

import "strings"

func compareManualDir(a, b manualDir) int {
	if a.Count != b.Count {
		return b.Count - a.Count
	}
	return strings.Compare(a.Path, b.Path)
}

func takeManualDirs(items []manualDir, limit int) []manualDir {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}
