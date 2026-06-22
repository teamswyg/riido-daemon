package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUploadsIncludeProductEvidenceWhenConfigured(t *testing.T) {
	dir := t.TempDir()
	product := filepath.Join(dir, "product.json")
	release := filepath.Join(dir, "release.json")
	lab := filepath.Join(dir, "lab.html")
	schedule := filepath.Join(dir, "schedule.json")
	infra := filepath.Join(dir, "infra.json")
	if err := os.WriteFile(product, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(lab, []byte("<html></html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(release, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(schedule, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(infra, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := uploadTestConfig(product, release, lab, schedule, infra)
	got := uploads(cfg, "20260622T000000Z")
	if len(got) != 16 {
		t.Fatalf("uploads=%d", len(got))
	}
	if got[6].target != "s3://bucket/daily/latest/ai-agent-product-acceptance.json" {
		t.Fatalf("product latest target=%q", got[6].target)
	}
	if got[8].target != "s3://bucket/daily/latest/local-release-acceptance.json" {
		t.Fatalf("release latest target=%q", got[8].target)
	}
	if got[10].target != "s3://bucket/daily/latest/contract-lab/index.html" {
		t.Fatalf("lab latest target=%q", got[10].target)
	}
	if got[12].target != "s3://bucket/daily/latest/local-qa-schedule.json" {
		t.Fatalf("schedule latest target=%q", got[12].target)
	}
	if got[14].target != "s3://bucket/daily/latest/local-qa-dashboard-infra-evidence.json" {
		t.Fatalf("infra latest target=%q", got[14].target)
	}
}

func uploadTestConfig(product, release, lab, schedule, infra string) config {
	provider := ".riido-local/provider.json"
	run := ".riido-local/run.json"
	dashboard := ".riido-local/index.html"
	screenshots := ".riido-local/screenshots"
	prefix := "s3://bucket/daily"
	return config{
		providerEvidence:   &provider,
		productEvidence:    &product,
		releaseEvidence:    &release,
		productLab:         &lab,
		scheduleEvidence:   &schedule,
		infraEvidence:      &infra,
		runEvidence:        &run,
		dashboardHTML:      &dashboard,
		productScreenshots: &screenshots,
		s3Prefix:           &prefix,
	}
}
