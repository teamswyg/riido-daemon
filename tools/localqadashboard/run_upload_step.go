package main

import "strings"

const (
	statusPassed = "passed"
	statusFailed = "failed"
)

func isUploadStep(id string) bool {
	return strings.HasPrefix(id, "upload-")
}
