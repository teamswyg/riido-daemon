package main

import "testing"

func TestUploadsIncludeProductEvidenceWhenConfigured(t *testing.T) {
	cfg := uploadTestConfig()
	got := uploads(cfg, "20260622T000000Z")
	if len(got) != 8 {
		t.Fatalf("uploads=%d", len(got))
	}
	if got[6].target != "s3://bucket/daily/latest/ai-agent-product-acceptance.json" {
		t.Fatalf("product latest target=%q", got[6].target)
	}
}

func uploadTestConfig() config {
	provider := ".riido-local/provider.json"
	product := ".riido-local/product.json"
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
