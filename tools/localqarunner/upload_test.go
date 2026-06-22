package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUploadsIncludeProductEvidenceWhenConfigured(t *testing.T) {
	dir := t.TempDir()
	product := filepath.Join(dir, "product.json")
	lab := filepath.Join(dir, "lab.html")
	schedule := filepath.Join(dir, "schedule.json")
	if err := os.WriteFile(product, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(lab, []byte("<html></html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(schedule, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := uploadTestConfig(product, lab, schedule)
	got := uploads(cfg, "20260622T000000Z")
	if len(got) != 12 {
		t.Fatalf("uploads=%d", len(got))
	}
	if got[6].target != "s3://bucket/daily/latest/ai-agent-product-acceptance.json" {
		t.Fatalf("product latest target=%q", got[6].target)
	}
	if got[8].target != "s3://bucket/daily/latest/contract-lab/index.html" {
		t.Fatalf("lab latest target=%q", got[8].target)
	}
	if got[10].target != "s3://bucket/daily/latest/local-qa-schedule.json" {
		t.Fatalf("schedule latest target=%q", got[10].target)
	}
}

func uploadTestConfig(product, lab, schedule string) config {
	provider := ".riido-local/provider.json"
	run := ".riido-local/run.json"
	dashboard := ".riido-local/index.html"
	screenshots := ".riido-local/screenshots"
	prefix := "s3://bucket/daily"
	return config{
		providerEvidence:   &provider,
		productEvidence:    &product,
		productLab:         &lab,
		scheduleEvidence:   &schedule,
		runEvidence:        &run,
		dashboardHTML:      &dashboard,
		productScreenshots: &screenshots,
		s3Prefix:           &prefix,
	}
}
