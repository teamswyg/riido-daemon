package main

import "testing"

func TestFinalDashboardUploadsSyncLatestAndStampedRunEvidence(t *testing.T) {
	cfg := uploadTestConfig("", "", ".riido-local/coverage.json", "", "", "")
	got := finalDashboardUploads(cfg, "20260622T000000Z", "s3://bucket/daily")
	if len(got) != 6 {
		t.Fatalf("uploads=%d", len(got))
	}
	if got[0].target != "s3://bucket/daily/latest/index.html" {
		t.Fatalf("dashboard latest target=%q", got[0].target)
	}
	if got[1].target != "s3://bucket/daily/latest/local-qa-run.json" {
		t.Fatalf("run latest target=%q", got[1].target)
	}
	if got[2].target != "s3://bucket/daily/latest/local-qa-coverage.json" {
		t.Fatalf("coverage latest target=%q", got[2].target)
	}
	if got[4].target != "s3://bucket/daily/20260622T000000Z/local-qa-run.json" {
		t.Fatalf("run stamped target=%q", got[4].target)
	}
}
