package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunEvidenceScenariosReportsS3Publish(t *testing.T) {
	path := filepath.Join(t.TempDir(), "run.json")
	body := `{"steps":[{"id":"upload-dashboard-html","status":"passed"}]}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := runEvidenceScenarios(path)
	if len(got) != 1 || got[0].ID != "local.qa.s3_publish" || got[0].Status != "passed" {
		t.Fatalf("scenarios=%+v", got)
	}
}
