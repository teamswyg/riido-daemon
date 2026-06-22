package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInfraEvidenceScenariosPassesPrivateS3Dashboard(t *testing.T) {
	path := filepath.Join(t.TempDir(), "infra.json")
	body := `{"terraform_managed":true,"public_access_blocked":true,"encryption_algorithm":"AES256","lifecycle_expire_days":30,"latest_index_observed":true,"latest_index_bytes":100,"latest_cache_control":"no-store","latest_object_sse":"AES256"}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := infraEvidenceScenarios(path)
	if len(got) != 1 || got[0].Status != statusPassed {
		t.Fatalf("scenarios=%+v", got)
	}
}

func TestInfraEvidenceScenariosFailsWhenBucketIsPublic(t *testing.T) {
	path := filepath.Join(t.TempDir(), "infra.json")
	body := `{"terraform_managed":true,"public_access_blocked":false,"encryption_algorithm":"AES256","lifecycle_expire_days":30,"latest_index_observed":true,"latest_index_bytes":100}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := infraEvidenceScenarios(path)
	if len(got) != 1 || got[0].Status != statusFailed {
		t.Fatalf("scenarios=%+v", got)
	}
}

func TestInfraEvidenceScenariosFailsWhenLatestObjectIsNotPinnedPrivate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "infra.json")
	body := `{"terraform_managed":true,"public_access_blocked":true,"encryption_algorithm":"AES256","lifecycle_expire_days":30,"latest_index_observed":true,"latest_index_bytes":100,"latest_cache_control":"max-age=3600","latest_object_sse":""}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := infraEvidenceScenarios(path)
	if len(got) != 1 || got[0].Status != statusFailed {
		t.Fatalf("scenarios=%+v", got)
	}
}
