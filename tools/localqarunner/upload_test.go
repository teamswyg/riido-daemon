package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUploadsIncludeProductEvidenceWhenConfigured(t *testing.T) {
	product := filepath.Join(t.TempDir(), "product.json")
	if err := os.WriteFile(product, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := uploadTestConfig(product)
	got := uploads(cfg, "20260622T000000Z")
	if len(got) != 8 {
		t.Fatalf("uploads=%d", len(got))
	}
	if got[6].target != "s3://bucket/daily/latest/ai-agent-product-acceptance.json" {
		t.Fatalf("product latest target=%q", got[6].target)
	}
}

func uploadTestConfig(product string) config {
	provider := ".riido-local/provider.json"
	run := ".riido-local/run.json"
	dashboard := ".riido-local/index.html"
	prefix := "s3://bucket/daily"
	return config{
		providerEvidence: &provider,
		productEvidence:  &product,
		runEvidence:      &run,
		dashboardHTML:    &dashboard,
		s3Prefix:         &prefix,
	}
}
