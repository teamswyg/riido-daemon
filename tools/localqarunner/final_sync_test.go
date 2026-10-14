package main

import "testing"

func TestFinalDashboardUploadsSyncLatestAndStampedRunEvidence(t *testing.T) {
	cfg := uploadTestConfig("", "", ".riido-local/coverage.json", "", "", "")
	got := finalDashboardUploads(cfg, "20260622T000000Z", "s3://bucket/daily")
	if len(got) != 4 {
		t.Fatalf("uploads=%d", len(got))
	}
	if got[0].target != "s3://bucket/daily/latest/index.html" {
		t.Fatalf("dashboard latest target=%q", got[0].target)
	}
	if got[1].target != "s3://bucket/daily/latest/local-qa-coverage.json" {
		t.Fatalf("coverage latest target=%q", got[1].target)
	}
	if got[3].target != "s3://bucket/daily/20260622T000000Z/local-qa-coverage.json" {
		t.Fatalf("coverage stamped target=%q", got[3].target)
	}
}

func TestFinalDashboardUploadsIncludeManualEvidenceWhenPresent(t *testing.T) {
	manual := writeUploadFixture(t, "manual.json", "{}")
	cfg := uploadTestConfig("", "", ".riido-local/coverage.json", "", "", "")
	cfg.manualEvidence = &manual
	got := finalDashboardUploads(cfg, "20260622T000000Z", "s3://bucket/daily")
	if len(got) != 6 {
		t.Fatalf("uploads=%d", len(got))
	}
	if got[4].target != "s3://bucket/daily/latest/manual-qa-evidence.json" {
		t.Fatalf("manual latest target=%q", got[4].target)
	}
	if got[5].target != "s3://bucket/daily/20260622T000000Z/manual-qa-evidence.json" {
		t.Fatalf("manual stamped target=%q", got[5].target)
	}
}

func TestFinalRunEvidenceUploadsSyncLatestAndStampedRunEvidence(t *testing.T) {
	cfg := uploadTestConfig("", "", ".riido-local/coverage.json", "", "", "")
	got := finalRunEvidenceUploads(cfg, "20260622T000000Z", "s3://bucket/daily")
	if len(got) != 2 {
		t.Fatalf("uploads=%d", len(got))
	}
	if got[0].target != "s3://bucket/daily/latest/local-qa-run.json" {
		t.Fatalf("run latest target=%q", got[0].target)
	}
	if got[1].target != "s3://bucket/daily/20260622T000000Z/local-qa-run.json" {
		t.Fatalf("run stamped target=%q", got[1].target)
	}
}

func TestSyncFinalDashboardArtifactsRecordsUploadSteps(t *testing.T) {
	cfg := uploadTestConfig("", "", ".riido-local/coverage.json", "", "", "")
	evidence := runEvidence{ObservedAt: "2026-06-22T00:00:00Z"}
	original := runFinalSyncStep
	t.Cleanup(func() { runFinalSyncStep = original })
	runFinalSyncStep = func(root, id, exe string, args ...string) stepEvidence {
		return stepEvidence{ID: id, Status: statusPassed, Command: exe}
	}
	if err := syncFinalDashboardArtifacts(".", cfg, &evidence); err != nil {
		t.Fatal(err)
	}
	if len(evidence.Steps) != 4 {
		t.Fatalf("steps=%d", len(evidence.Steps))
	}
	if evidence.Steps[0].ID != "upload-dashboard-html-final" {
		t.Fatalf("first step=%+v", evidence.Steps[0])
	}
}
