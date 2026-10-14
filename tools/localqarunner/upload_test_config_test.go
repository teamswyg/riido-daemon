package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeUploadFixture(t *testing.T, name, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func uploadTestConfig(product, release, coverage, lab, schedule, infra string) config {
	provider := ".riido-local/provider.json"
	run := ".riido-local/run.json"
	dashboard := ".riido-local/index.html"
	manual := ""
	domainCache := ""
	screenshots := ".riido-local/screenshots"
	prefix := "s3://bucket/daily"
	return config{
		providerEvidence:   &provider,
		productEvidence:    &product,
		releaseEvidence:    &release,
		coverageEvidence:   &coverage,
		manualEvidence:     &manual,
		domainCache:        &domainCache,
		productLab:         &lab,
		scheduleEvidence:   &schedule,
		infraEvidence:      &infra,
		runEvidence:        &run,
		dashboardHTML:      &dashboard,
		productScreenshots: &screenshots,
		s3Prefix:           &prefix,
	}
}
