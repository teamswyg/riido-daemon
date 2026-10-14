package main

import (
	pathmatch "path"
	"strings"
)

func countUploadedEvidenceOut(evidenceOut, uploadPaths []string) int {
	count := 0
	for _, path := range evidenceOut {
		if evidenceOutUploaded(path, uploadPaths) {
			count++
		}
	}
	return count
}

func missingEvidenceUploads(evidenceOut, uploadPaths []string) []string {
	var missing []string
	for _, path := range evidenceOut {
		if !evidenceOutUploaded(path, uploadPaths) {
			missing = append(missing, path)
		}
	}
	return uniqueStrings(missing)
}

func evidenceOutUploaded(path string, uploadPaths []string) bool {
	for _, upload := range uploadPaths {
		if upload == path ||
			uploadCoversVariableEvidenceOut(upload, path) ||
			uploadCoversGlob(upload, path) ||
			uploadCoversDirectory(upload, path) {
			return true
		}
	}
	return false
}

func uploadCoversVariableEvidenceOut(upload, path string) bool {
	start := strings.Index(path, "${")
	end := strings.Index(path, "}")
	if start < 0 || end < start {
		return false
	}
	return strings.HasPrefix(upload, path[:start]) && strings.HasSuffix(upload, path[end+1:])
}

func uploadCoversGlob(upload, evidencePath string) bool {
	if !strings.ContainsAny(upload, "*?[") {
		return false
	}
	ok, err := pathmatch.Match(upload, evidencePath)
	return err == nil && ok
}

func uploadCoversDirectory(upload, evidencePath string) bool {
	upload = strings.TrimRight(upload, "/")
	if upload == "" || strings.ContainsAny(upload, "*?[") {
		return false
	}
	return strings.HasPrefix(evidencePath, upload+"/")
}
