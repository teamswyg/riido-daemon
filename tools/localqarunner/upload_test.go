package main

import "testing"

func TestUploadsIncludeProductEvidenceWhenConfigured(t *testing.T) {
	product := writeUploadFixture(t, "product.json", "{}")
	release := writeUploadFixture(t, "release.json", "{}")
	coverage := writeUploadFixture(t, "coverage.json", "{}")
	lab := writeUploadFixture(t, "lab.html", "<html></html>")
	schedule := writeUploadFixture(t, "schedule.json", "{}")
	infra := writeUploadFixture(t, "infra.json", "{}")
	cfg := uploadTestConfig(product, release, coverage, lab, schedule, infra)
	got := uploads(cfg, "20260622T000000Z")
	if len(got) != 18 {
		t.Fatalf("uploads=%d", len(got))
	}
	if got[6].target != "s3://bucket/daily/latest/ai-agent-product-acceptance.json" {
		t.Fatalf("product latest target=%q", got[6].target)
	}
	if got[8].target != "s3://bucket/daily/latest/local-release-acceptance.json" {
		t.Fatalf("release latest target=%q", got[8].target)
	}
	if got[10].target != "s3://bucket/daily/latest/local-qa-coverage.json" {
		t.Fatalf("coverage latest target=%q", got[10].target)
	}
	if got[12].target != "s3://bucket/daily/latest/contract-lab/index.html" {
		t.Fatalf("lab latest target=%q", got[12].target)
	}
	if got[14].target != "s3://bucket/daily/latest/local-qa-schedule.json" {
		t.Fatalf("schedule latest target=%q", got[14].target)
	}
	if got[16].target != "s3://bucket/daily/latest/local-qa-dashboard-infra-evidence.json" {
		t.Fatalf("infra latest target=%q", got[16].target)
	}
}

func TestUploadsIncludeManualEvidenceWhenPresent(t *testing.T) {
	manual := writeUploadFixture(t, "manual.json", "{}")
	cfg := uploadTestConfig("", "", "", "", "", "")
	cfg.manualEvidence = &manual
	got := uploads(cfg, "20260622T000000Z")
	if len(got) != 8 {
		t.Fatalf("uploads=%d", len(got))
	}
	if got[6].target != "s3://bucket/daily/latest/manual-qa-evidence.json" {
		t.Fatalf("manual latest target=%q", got[6].target)
	}
	if got[7].target != "s3://bucket/daily/20260622T000000Z/manual-qa-evidence.json" {
		t.Fatalf("manual stamped target=%q", got[7].target)
	}
}

func TestUploadsIncludeDomainCacheWhenPresent(t *testing.T) {
	domainCache := writeUploadFixture(t, "domain-cache.json", "{}")
	cfg := uploadTestConfig("", "", "", "", "", "")
	cfg.domainCache = &domainCache
	got := uploads(cfg, "20260622T000000Z")
	if len(got) != 8 {
		t.Fatalf("uploads=%d", len(got))
	}
	if got[6].target != "s3://bucket/daily/latest/domain-fixture-journey-cache.json" {
		t.Fatalf("domain latest target=%q", got[6].target)
	}
	if got[7].target != "s3://bucket/daily/20260622T000000Z/domain-fixture-journey-cache.json" {
		t.Fatalf("domain stamped target=%q", got[7].target)
	}
}
